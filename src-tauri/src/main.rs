#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use std::sync::{atomic::{AtomicBool, Ordering}, Mutex};
use tauri::Manager;
use tauri_plugin_shell::process::CommandChild;

#[derive(Default)]
struct Server {
    children: Mutex<Vec<CommandChild>>,
    stopping: AtomicBool,
}

impl Server {
    fn add(&self, child: CommandChild) {
        let mut children = self.children.lock().unwrap();
        if self.stopping.load(Ordering::SeqCst) {
            let _ = child.kill();
        } else {
            children.push(child);
        }
    }

    fn stop(&self) {
        self.stopping.store(true, Ordering::SeqCst);
        if let Ok(mut children) = self.children.lock() {
            for child in children.drain(..).rev() {
                let _ = child.kill();
            }
        }
    }
}

#[cfg(not(dev))]
fn forward_events(
    mut receiver: tauri::async_runtime::Receiver<tauri_plugin_shell::process::CommandEvent>,
    sender: tauri::async_runtime::Sender<(&'static str, tauri_plugin_shell::process::CommandEvent)>,
    name: &'static str,
) {
    tauri::async_runtime::spawn(async move {
        while let Some(event) = receiver.recv().await {
            if sender.send((name, event)).await.is_err() {
                return;
            }
        }
        let _ = sender.send((name, tauri_plugin_shell::process::CommandEvent::Error("服务管道已关闭".into()))).await;
    });
}

#[cfg(not(dev))]
async fn start_server(app: tauri::AppHandle) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    use flate2::read::GzDecoder;
    use std::{fs, io::Write, sync::mpsc, time::Duration};
    use tauri_plugin_shell::{process::CommandEvent, ShellExt};

    #[derive(serde::Deserialize)]
    #[serde(rename_all = "camelCase")]
    struct BuildInfo {
        version: String,
        build_id: String,
    }

    let resources = app.path().resource_dir()?.join("resources");
    let build: BuildInfo = serde_json::from_slice(&fs::read(resources.join("desktop-build.json"))?)?;
    if !build.build_id.chars().all(|c| c.is_ascii_alphanumeric() || c == '-' || c == '_') {
        return Err("安装包中的构建标识无效".into());
    }
    // Next 的缓存必须可写；按版本和构建隔离，重启复用已经解压的资源。
    let runtime = app.path().app_cache_dir()?.join("server").join(build.version).join(build.build_id);
    if !runtime.join(".ready").exists() {
        fs::create_dir_all(&runtime)?;
        tar::Archive::new(GzDecoder::new(fs::File::open(resources.join("server.tar.gz"))?)).unpack(&runtime)?;
        fs::write(runtime.join(".ready"), b"ready")?;
    }
    let log_dir = app.path().app_log_dir()?;
    fs::create_dir_all(&log_dir)?;
    let log_path = log_dir.join("next-server.log");
    let mut log = fs::File::create(&log_path)?;

    let (sender, mut events) = tauri::async_runtime::channel(64);
    let (cancel_timeout, timeout) = mpsc::channel::<()>();
    let mut cancel_timeout = Some(cancel_timeout);
    let timeout_sender = sender.clone();
    std::thread::spawn(move || {
        if matches!(timeout.recv_timeout(Duration::from_secs(60)), Err(mpsc::RecvTimeoutError::Timeout)) {
            let _ = timeout_sender.blocking_send(("桌面", CommandEvent::Error("服务未能在 60 秒内就绪".into())));
        }
    });

    let mut api_base_url = std::env::var("API_BASE_URL").ok().map(|value| value.trim().to_string()).filter(|value| !value.is_empty());
    if api_base_url.is_none() {
        let data_dir = app.path().app_data_dir()?;
        fs::create_dir_all(&data_dir)?;
        let (api_events, api_child) = app.shell().sidecar("server")?
            .env("MOHUA_DESKTOP", "1")
            .env("GIN_MODE", "release")
            .env("STORAGE_DRIVER", "sqlite")
            .env("DATABASE_DSN", data_dir.join("infinite-canvas.db"))
            .env("AI_LOG_DIR", data_dir.join("logs").join("ai-calls"))
            .current_dir(&data_dir)
            .spawn()?;
        app.state::<Server>().add(api_child);
        forward_events(api_events, sender.clone(), "Go");
    }

    let mut next_started = false;
    loop {
        if let Some(url) = api_base_url.take() {
            let (next_events, child) = app.shell().sidecar("node")?
                .args([runtime.join("desktop-server.cjs")])
                .current_dir(&runtime)
                .env("NODE_ENV", "production")
                .env("NEXT_TELEMETRY_DISABLED", "1")
                .env("API_BASE_URL", url)
                .spawn()?;
            app.state::<Server>().add(child);
            forward_events(next_events, sender.clone(), "Next");
            next_started = true;
        }
        let Some((name, event)) = events.recv().await else { break };
        match event {
            CommandEvent::Stdout(bytes) | CommandEvent::Stderr(bytes) => {
                let line = String::from_utf8_lossy(&bytes);
                let _ = writeln!(log, "[{name}] {}", line.trim_end());
                if name == "Go" && !next_started {
                    if let Some(url) = line.trim().strip_prefix("MOHUA_API_READY=") {
                        api_base_url = Some(url.to_string());
                    }
                }
                if name == "Next" && line.trim() == "MOHUA_DESKTOP_READY" {
                    drop(cancel_timeout.take());
                    let window = app.get_webview_window("main").ok_or("主窗口不存在")?;
                    window.navigate("http://127.0.0.1:39217".parse()?)?;
                }
            }
            CommandEvent::Error(error) => return Err(format!("{name} 服务错误：{error}。日志：{}", log_path.display()).into()),
            CommandEvent::Terminated(status) => {
                return Err(format!("{name} 服务已停止（退出码 {:?}）。日志：{}", status.code, log_path.display()).into());
            }
            _ => {}
        }
    }
    Err(format!("本地服务连接已关闭。日志：{}", log_path.display()).into())
}

fn main() {
    tauri::Builder::default()
        .plugin(tauri_plugin_single_instance::init(|app, _, _| {
            if let Some(window) = app.get_webview_window("main") {
                let _ = window.unminimize();
                let _ = window.show();
                let _ = window.set_focus();
            }
        }))
        .plugin(tauri_plugin_shell::init())
        .plugin(tauri_plugin_dialog::init())
        .manage(Server::default())
        .setup(|app| {
            #[cfg(not(dev))]
            {
                use tauri_plugin_dialog::{DialogExt, MessageDialogKind};
                let handle = app.handle().clone();
                tauri::async_runtime::spawn(async move {
                    if let Err(error) = start_server(handle.clone()).await {
                        if handle.state::<Server>().stopping.load(Ordering::SeqCst) {
                            return;
                        }
                        handle.state::<Server>().stop();
                        let exit_handle = handle.clone();
                        handle.dialog().message(format!("桌面端启动或运行失败：{error}"))
                            .title("墨画画布")
                            .kind(MessageDialogKind::Error)
                            .show(move |_| exit_handle.exit(1));
                    }
                });
            }
            #[cfg(dev)]
            let _ = app;
            Ok(())
        })
        .build(tauri::generate_context!())
        .expect("无法创建桌面应用")
        .run(|app, event| {
            if matches!(event, tauri::RunEvent::Exit) {
                app.state::<Server>().stop();
            }
        });
}
