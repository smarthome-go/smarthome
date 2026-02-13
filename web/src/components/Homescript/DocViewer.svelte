<script lang="ts">
    import type { DocEntry, DocModuleGroup, DocsContents } from './docs'
    import { getDocEntries, groupDocsByModule } from './docs'
    import Textfield from '@smui/textfield'

    export let contents: DocsContents = []

    let entries: DocEntry[] = []
    let filteredEntries: DocEntry[] = []
    let modules: DocModuleGroup[] = []
    let activeEntries: DocEntry[] = []
    let activeModule = ''
    let search = ''
    let searchQuery = ''
    let emptyMessage = 'No docs available.'

    $: entries = getDocEntries(contents)
    $: searchQuery = search.trim().toLowerCase()
    $: filteredEntries =
        searchQuery.length === 0
            ? entries
            : entries.filter(entry => {
                  return (
                      entry.name.toLowerCase().includes(searchQuery) ||
                      entry.module.toLowerCase().includes(searchQuery)
                  )
              })
    $: modules = groupDocsByModule(filteredEntries)
    $: emptyMessage = entries.length === 0 ? 'No docs available.' : 'No matches.'
    $: if (modules.length === 0) {
        activeModule = ''
    } else if (!activeModule || !modules.some(module => module.module === activeModule)) {
        activeModule = modules[0].module
    }
    $: activeEntries = activeModule
        ? modules.find(module => module.module === activeModule)?.entries ?? []
        : []

    function formatKind(kind: string): string {
        return kind.toUpperCase()
    }
</script>

<div class="doc-viewer">
    <div class="doc-viewer__toolbar">
        <Textfield
            bind:value={search}
            label="Search functions / modules"
            style="width: 100%;"
            helperLine$style="display: none;"
        />
    </div>
    <div class="doc-viewer__layout">
        <aside class="doc-viewer__modules">
            {#if modules.length === 0}
                <div class="doc-viewer__empty">{emptyMessage}</div>
            {:else}
                {#each modules as module (module.module)}
                    <button
                        class="doc-module-button"
                        class:active={module.module === activeModule}
                        on:click={() => (activeModule = module.module)}
                        on:focus={() => (activeModule = module.module)}
                        type="button"
                        aria-pressed={module.module === activeModule}
                    >
                        <span class="doc-module-button__name">{module.module}</span>
                        <span class="doc-module-button__count">{module.entries.length}</span>
                    </button>
                {/each}
            {/if}
        </aside>
        <section class="doc-viewer__content">
            {#if activeModule === ''}
                <div class="doc-viewer__empty">{emptyMessage}</div>
            {:else}
                <header class="doc-module__header">
                    <h3 class="doc-module__title">{activeModule}</h3>
                    <span class="doc-module__count">{activeEntries.length}</span>
                </header>
                <div class="doc-module__entries">
                    {#each activeEntries as entry (entry.module + ':' + entry.name + ':' + entry.kind)}
                        <article class="doc-entry">
                            <div class="doc-entry__title">
                                <span class="doc-entry__name">{entry.name}</span>
                                <span class="doc-entry__kind">{formatKind(entry.kind)}</span>
                            </div>
                            {#if entry.signature}
                                <pre class="doc-entry__signature">{entry.signature}</pre>
                            {:else if entry.triggerSignature || entry.callbackSignature}
                                <div class="doc-entry__signature-group">
                                    {#if entry.triggerSignature}
                                        <div class="doc-entry__signature-label">trigger</div>
                                        <pre class="doc-entry__signature">{entry.triggerSignature}</pre>
                                    {/if}
                                    {#if entry.callbackSignature}
                                        <div class="doc-entry__signature-label">callback</div>
                                        <pre class="doc-entry__signature">{entry.callbackSignature}</pre>
                                    {/if}
                                </div>
                            {:else}
                                <div class="doc-entry__empty">No signature</div>
                            {/if}
                        </article>
                    {/each}
                </div>
            {/if}
        </section>
    </div>
</div>

<style lang="scss">
    .doc-viewer {
        display: flex;
        flex-direction: column;
        gap: 1rem;
        padding: 0.75rem;
    }

    .doc-viewer__toolbar {
        position: sticky;
        top: 0;
        z-index: 2;
        background: var(--clr-height-0-1);
        padding-bottom: 0.25rem;
    }

    .doc-viewer__layout {
        display: grid;
        grid-template-columns: minmax(180px, 240px) minmax(0, 1fr);
        gap: 0.75rem;
        min-height: 24rem;
        max-height: 68vh;
        overflow: hidden;
    }

    .doc-viewer__modules {
        display: flex;
        flex-direction: column;
        gap: 0.4rem;
        padding: 0.5rem;
        border-radius: 0.6rem;
        border: 1px solid var(--clr-height-0-3);
        background: var(--clr-height-0-2);
        overflow: auto;
    }

    .doc-viewer__content {
        display: flex;
        flex-direction: column;
        border-radius: 0.6rem;
        border: 1px solid var(--clr-height-0-3);
        background: var(--clr-height-0-2);
        overflow: auto;
    }

    .doc-viewer__empty {
        padding: 1rem;
        border-radius: 0.5rem;
        background: var(--clr-height-0-2);
        color: var(--clr-text-disabled);
        text-align: center;
    }

    .doc-module-button {
        border: 1px solid transparent;
        background: var(--clr-height-0-1);
        color: var(--clr-text);
        padding: 0.45rem 0.6rem;
        border-radius: 0.45rem;
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 0.6rem;
        cursor: pointer;
        text-align: left;
    }

    .doc-module-button:hover,
    .doc-module-button:focus-visible {
        border-color: var(--clr-height-1-3);
        background: var(--clr-height-1-3);
        outline: none;
    }

    .doc-module-button.active {
        border-color: var(--clr-primary-light);
        background: var(--clr-height-1-3);
    }

    .doc-module-button__name {
        font-size: 0.9rem;
    }

    .doc-module-button__count {
        font-size: 0.75rem;
        color: var(--clr-text-hint);
        background: var(--clr-height-0-3);
        border-radius: 999px;
        padding: 0.15rem 0.45rem;
    }

    .doc-module__header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0.75rem 1rem;
        background: var(--clr-height-1-3);
        border-bottom: 1px solid var(--clr-height-0-3);
    }

    .doc-module__title {
        margin: 0;
        font-size: 1rem;
        letter-spacing: 0.02em;
    }

    .doc-module__count {
        font-size: 0.8rem;
        color: var(--clr-text-hint);
        background: var(--clr-height-0-3);
        border-radius: 999px;
        padding: 0.15rem 0.55rem;
    }

    .doc-module__entries {
        display: flex;
        flex-direction: column;
    }

    .doc-entry {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        padding: 0.85rem 1rem;
        border-top: 1px solid var(--clr-height-0-3);
    }

    .doc-entry:first-child {
        border-top: none;
    }

    .doc-entry__title {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 0.75rem;
    }

    .doc-entry__name {
        font-weight: 600;
    }

    .doc-entry__kind {
        font-size: 0.7rem;
        letter-spacing: 0.08em;
        color: var(--clr-text-hint);
        background: var(--clr-height-0-3);
        border-radius: 0.3rem;
        padding: 0.2rem 0.4rem;
    }

    .doc-entry__signature-group {
        display: grid;
        gap: 0.4rem;
    }

    .doc-entry__signature-label {
        font-size: 0.7rem;
        text-transform: uppercase;
        letter-spacing: 0.08em;
        color: var(--clr-text-hint);
    }

    .doc-entry__signature {
        margin: 0;
        padding: 0.55rem 0.7rem;
        background: var(--clr-height-0-1);
        border-radius: 0.4rem;
        border: 1px solid var(--clr-height-0-3);
        font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, "Liberation Mono", monospace;
        font-size: 0.85rem;
        line-height: 1.3;
        white-space: pre-wrap;
    }

    .doc-entry__empty {
        color: var(--clr-text-hint);
        font-size: 0.85rem;
    }

    @media (max-width: 900px) {
        .doc-viewer__layout {
            grid-template-columns: 1fr;
            max-height: none;
        }
    }
</style>
