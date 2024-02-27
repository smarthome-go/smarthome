const hmsEditorURL = "/homescript/editor"

export function hmsEditorURLForId(id: string): string {
    const sanitizedId = encodeURIComponent(id)
    return `${hmsEditorURL}?id=${sanitizedId}`
}
