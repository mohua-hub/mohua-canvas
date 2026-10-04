export const imageResolutionOptions = [
    { value: "1k", label: "1k" },
    { value: "2k", label: "2k" },
    { value: "4k", label: "4k" },
] as const;

const resolutionBase = { "1k": 1024, "2k": 2048, "4k": 2880 } as const;

export function normalizeImageResolution(value: string) {
    const normalized = value.trim().toLowerCase();
    return normalized in resolutionBase ? normalized as keyof typeof resolutionBase : "1k";
}

export function imageSizeForResolution(size: string, resolution: string) {
    const value = size.trim().toLowerCase();
    if (!value || value === "auto") return value || undefined;
    const match = value.match(/^(\d+)[x:](\d+)$/);
    if (!match) return undefined;
    const width = Number(match[1]);
    const height = Number(match[2]);
    if (!width || !height) return undefined;
    const divisor = greatestCommonDivisor(width, height);
    const ratioWidth = width / divisor;
    const ratioHeight = height / divisor;
    const base = resolutionBase[normalizeImageResolution(resolution)];
    const unit = Math.round(Math.sqrt((base * base) / (ratioWidth * ratioHeight)) / 16) * 16;
    return `${ratioWidth * unit}x${ratioHeight * unit}`;
}

function greatestCommonDivisor(a: number, b: number) {
    while (b) [a, b] = [b, a % b];
    return a;
}
