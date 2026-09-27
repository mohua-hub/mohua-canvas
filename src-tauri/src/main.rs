#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use std::sync::Mutex;
use tauri::Manager;
use tauri_plugin_shell::process::CommandChild;

#[derive(Default)]
struct Server(Mutex<Option<CommandChild>>);

impl Server {
    fn stop(&self) {
        if let Ok(mut server) = self.0.lock() {
            if let Some(child) = server.take() {
                let _ = child.kill();
            }
        }
    }
}

#[cfg(not(dev))]
async fn start_server(app: tauri::AppHandle) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    use flate2::read::GzDecoder;
    use std::{fs, io::Write};
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
    let (mut events, child) = app.shell().sidecar("node")?
        .args([runtime.join("desktop-server.cjs")])
        .current_dir(&runtime)
        .env("NODE_ENV", "production")
        .env("NEXT_TELEMETRY_DISABLED", "1")
        .spawn()?;
    *app.state::<Server>().0.lock().unwrap() = Some(child);

    while let Some(event) = events.recv().await {
        match event {
            CommandEvent::Stdout(bytes) | CommandEvent::Stderr(bytes) => {
                let _ = log.write_all(&bytes);
                let _ = log.write_all(b"\n");
                if String::from_utf8_lossy(&bytes).trim() == "MOHUA_DESKTOP_READY" {
                    let window = app.get_webview_window("main").ok_or("主窗口不存在")?;
                    window.navigate("http://127.0.0.1:39217".parse()?)?;
                }
            }
            CommandEvent::Error(error) => return Err(error.into()),
            CommandEvent::Terminated(status) => {
                return Err(format!("本地服务已停止（退出码 {:?}）。日志：{}", status.code, log_path.display()).into());
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
