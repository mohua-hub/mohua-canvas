import localforage from "localforage";

import type { WorkflowChannelData, WorkflowEntry } from "@/lib/workflow-channel";

const store = localforage.createInstance({ name: "infinite-canvas", storeName: "workflow_channels" });
const storageKey = "default";
let writeQueue = Promise.resolve();

function queueWrite(write: () => Promise<void>) {
    const next = writeQueue.then(write, write);
    writeQueue = next.catch(() => undefined);
    return next;
}

async function readChannels() {
    return (await store.getItem<WorkflowChannelData[]>(storageKey)) || [];
}

export async function readWorkflowChannel(protocol: WorkflowChannelData["protocol"], channelId: string): Promise<WorkflowEntry[]> {
	await writeQueue;
	const channels = await readChannels();
    return channels.find((channel) => channel.protocol === protocol && channel.channelId === channelId)?.workflows || [];
}

export async function saveWorkflowChannel(protocol: WorkflowChannelData["protocol"], channelId: string, workflows: WorkflowEntry[]) {
    await queueWrite(async () => {
        const channels = await readChannels();
        const index = channels.findIndex((channel) => channel.protocol === protocol && channel.channelId === channelId);
        const updatedChannels = [...channels];
        const channel = { protocol, channelId, workflows };
        if (index >= 0) updatedChannels[index] = channel;
        else updatedChannels.push(channel);
        await store.setItem(storageKey, updatedChannels);
    });
}

export async function listWorkflowChannels(): Promise<WorkflowChannelData[]> {
	await writeQueue;
	return readChannels();
}

export async function replaceWorkflowChannels(channels: WorkflowChannelData[]) {
    await queueWrite(() => store.setItem(storageKey, channels).then(() => undefined));
}
