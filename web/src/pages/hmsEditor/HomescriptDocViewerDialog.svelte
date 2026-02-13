<script lang="ts">
    import Dialog, { Actions, Content, Header, Title } from '@smui/dialog'
    import IconButton from '@smui/icon-button'
    import Button, { Label } from '@smui/button'
    import { onMount } from 'svelte'
    import Progress from '../../components/Progress.svelte'
    import { createSnackbar } from '../../global'
    import DocViewer from '../../components/Homescript/DocViewer.svelte'
    import { fetchDocs, type DocsResponse } from '../../components/Homescript/docs'

    export let open = false

    let loading = false
    let docs: DocsResponse | null = null
    let errorMessage = ''

    async function loadDocs() {
        loading = true
        errorMessage = ''
        try {
            docs = await fetchDocs()
        } catch (err) {
            errorMessage = err instanceof Error ? err.message : String(err)
            docs = null
            $createSnackbar(`Could not load Homescript docs: ${errorMessage}`)
        }
        loading = false
    }

    onMount(() => {
        loadDocs()
    })

    $: if (open && !loading && docs === null && errorMessage === '') {
        loadDocs()
    }
</script>

<Dialog
    bind:open
    fullscreen
    class="homescript-doc-dialog"
    aria-labelledby="docs-title"
    aria-describedby="docs-content"
>
    <Header>
        <Title id="docs-title">Homescript Docs</Title>
        <div class="doc-dialog__header-actions">
            <IconButton
                class="material-icons"
                title="Refresh"
                disabled={loading}
                on:click={loadDocs}
            >
                refresh
            </IconButton>
            <IconButton action="close" class="material-icons">close</IconButton>
        </div>
    </Header>
    <Content id="docs-content">
        <Progress type="linear" bind:loading />
        {#if errorMessage}
            <div class="doc-dialog__error">{errorMessage}</div>
        {/if}
        <DocViewer contents={docs} />
    </Content>
    <Actions>
        <Button on:click={() => (open = false)}>
            <Label>Close</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
    @use '../../_mixins.scss' as *;

    .doc-dialog__header-actions {
        display: flex;
        align-items: center;
        gap: 0.35rem;
    }

    :global(.homescript-doc-dialog .mdc-dialog__surface) {
        display: flex;
        flex-direction: column;
        height: 80vh;
        max-height: 80vh;
    }

    :global(.homescript-doc-dialog .mdc-dialog__content) {
        flex: 1;
        overflow: auto;
    }

    .doc-dialog__error {
        margin: 0.75rem 0 0.5rem;
        padding: 0.6rem 0.8rem;
        border-radius: 0.45rem;
        background: var(--clr-height-0-2);
        color: var(--clr-error);
        border: 1px solid var(--clr-height-0-3);
        font-size: 0.85rem;
    }

    :global(.homescript-doc-dialog .mdc-dialog__content) {
        padding: 0.5rem 0.75rem 1rem;
    }

    @include mobile {
        :global(.homescript-doc-dialog .mdc-dialog__surface) {
            height: 90vh;
            max-height: 90vh;
        }

        :global(.homescript-doc-dialog .mdc-dialog__content) {
            padding: 0.5rem 0.4rem 1rem;
        }
    }
</style>
