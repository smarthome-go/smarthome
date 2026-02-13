export type DocKind = 'normal' | 'type' | 'trigger' | 'template'

export interface DocEntry {
    module: string
    name: string
    kind: DocKind
    signature?: string
    triggerSignature?: string
    callbackSignature?: string
}

export interface DocsResponse {
    entries: DocEntry[]
}

export type DocsContents = DocsResponse | DocEntry[] | null | undefined

export interface DocModuleGroup {
    module: string
    entries: DocEntry[]
}

export function getDocEntries(contents: DocsContents): DocEntry[] {
    if (!contents) return []
    if (Array.isArray(contents)) return contents
    if (Array.isArray(contents.entries)) return contents.entries
    return []
}

export function groupDocsByModule(entries: DocEntry[]): DocModuleGroup[] {
    const modules = new Map<string, DocEntry[]>()

    for (const entry of entries) {
        const existing = modules.get(entry.module)
        if (existing) {
            existing.push(entry)
        } else {
            modules.set(entry.module, [entry])
        }
    }

    return [...modules.entries()]
        .map(([module, moduleEntries]) => ({
            module,
            entries: [...moduleEntries].sort((a, b) => a.name.localeCompare(b.name)),
        }))
        .sort((a, b) => a.module.localeCompare(b.module))
}

export async function fetchDocs(): Promise<DocsResponse> {
    const res = await fetch('/api/homescript/docs')
    if (!res.ok) throw Error(`Docs request failed: ${res.status}`)
    const data = await res.json()
    if (data?.success !== undefined && !data.success)
        throw Error(data.error ?? data.message ?? 'Docs request failed')
    return data as DocsResponse
}
