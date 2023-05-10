<script lang="ts">
    import Button, { Label } from '@smui/button'
    import Checkbox from '@smui/checkbox'
    import Dialog, { Actions, Content, Header, Title, InitialFocus } from '@smui/dialog'
    import FormField from '@smui/form-field'
    import Progress from '../../../components/Progress.svelte'
    import { createSnackbar } from '../../../global'

    export let open = false
    let loading = false

    let includeProfilePictures = false
    let includeCache = false

    function downloadTextFile(filename: string, content: string) {
        const temp = document.createElement('a')
        temp.href = 'data:text/plain;charset=utf-8,' + encodeURIComponent(content)
        temp.download = filename
        temp.style.display = 'Smarthome Configuration Export'
        document.body.appendChild(temp)
        temp.click()
        document.body.removeChild(temp)
    }

    async function exportConfig() {
        loading = true
        try {
            const res = await (
                await fetch('/api/system/config/export', {
                    method: 'POST',
                    body: JSON.stringify({
                        includeProfilePictures: includeProfilePictures,
                        includeCacheData: includeCache,
                    }),
                })
            ).json()
            if (res.success != undefined && !res.success) throw Error(res.error)
            // Download the fetched configuration
            downloadTextFile(
                `${window.location.hostname}_${new Date().toISOString()}_smarthome_export.json`,
                JSON.stringify(res, null, '\t'),
            )
            open = false
        } catch (err) {
            $createSnackbar(`Failed to export system configuration: ${err}`)
        }
        loading = false
    }
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <Header>
        <Title id="title">Configuration Export</Title>
    </Header>
    <Content id="content">
        <Progress bind:loading />
        <div id="confirm">
            <div class="list warn">
                <span>Before you continue</span>
                <ul>
                    <li>
                        Exporting will <strong>expose all data</strong> stored on this server.
                    </li>
                    <li>Store the export carefully, it contains sensitive information.</li>
                    <li>Exporting may take some time.</li>
                </ul>
            </div>
            <FormField>
                <Checkbox bind:checked={includeProfilePictures} />
                <span slot="label">include profile pictures</span>
            </FormField>
            <br>
            <FormField>
                <Checkbox bind:checked={includeCache} />
                <span slot="label">include cache data</span>
            </FormField>
            <Button id="export-button" variant="raised" on:click={exportConfig}>
                <Label>Export</Label>
            </Button>
        </div>
    </Content>
    <Actions>
        <Button
            defaultAction
            use={[InitialFocus]}
            on:click={() => {
                open = false
            }}
        >
            <Label>Cancel</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
    @use '../../../mixins' as *;

    #confirm {
        width: 30rem;

        @include mobile {
            width: auto;
        }
    }

    .list {
        margin-bottom: 0.5rem;

        &.warn {
            span {
                color: var(--clr-error);
            }
        }

        span {
            font-weight: bold;
            color: var(--clr-primary);
        }

        ul {
            padding: 0 1rem;
            margin: 0;
            margin-top: 0.125rem;
        }
    }

    :global #export-button {
        margin-top: 0.8rem;
        display: block;
    }
</style>
