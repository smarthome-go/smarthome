<script lang="ts">
    import { Icon, Label } from '@smui/button'
    import Button from '@smui/button/src/Button.svelte'
    import Progress from '../../components/Progress.svelte'
    import { createSnackbar } from '../../global'
    import ConfigImport from './dialogs/ConfigImport.svelte'
    import FactoryReset from './dialogs/FactoryReset.svelte'

    let importConfigOpen = false
    let factoryOpen = false

    let loading = false

    function downloadTextFile(filename: string, content: string) {
        const temp = document.createElement('a')
        temp.href = 'data:text/plain;charset=utf-8,' + encodeURIComponent(content)
        temp.download = filename
        temp.style.display = 'Configuration Export'
        document.body.appendChild(temp)
        temp.click()
        document.body.removeChild(temp)
    }

    async function exportConfig() {
        loading = true
        try {
            const res = await (await fetch('/api/system/config/export')).json()
            if (res.success != undefined && !res.success) throw Error(res.error)
            // Download the fetched configuration
            downloadTextFile(
                `${window.location.hostname}_${new Date().toISOString()}_smarthome_export.json`,
                JSON.stringify(res, null, '\t'),
            )
        } catch (err) {
            $createSnackbar(`Failed to export system configuration: ${err}`)
        }
        loading = false
    }
</script>

<ConfigImport bind:open={importConfigOpen} />
<FactoryReset bind:open={factoryOpen} />

<Progress bind:loading />

<h6>Export / Import</h6>
<div class="container">
    <div class="container__export mdc-elevation--z3">
        <div class="container__export__description ">
            <span class="text-hint">
                Will export nearly all configured settings of this server.
                <br />
                <strong style="color: var(--clr-warn);">Warning!</strong>
                Store this file securely. It contains passwords and sensitive information.
            </span>
        </div>
        <Button on:click={exportConfig} variant="raised">
            <Label>Export</Label>
            <Icon class="material-icons">file_download</Icon>
        </Button>
    </div>
    <div class="container__import mdc-elevation--z3">
        <div class="container__import__description">
            <span class="text-hint">
                Will import settings of another server.
                <br />
                <strong style="color: var(--clr-warn);">Warning!</strong>
                Only useful for fresh instances: will erase all data.
            </span>
        </div>
        <div class="container__import__buttons">
            <Button on:click={() => (importConfigOpen = true)} variant="outlined">
                <Label>Import</Label>
                <Icon class="material-icons">file_upload</Icon>
            </Button>
            <Button on:click={() => (factoryOpen = true)} variant="outlined">
                <Label>Factory</Label>
                <Icon class="material-icons">restart_alt</Icon>
            </Button>
        </div>
    </div>
</div>

<style lang="scss">
    @use '../../mixins' as *;

    h6 {
        margin: 0;
        color: var(--clr-text-hint);
        font-size: 1rem;

        @include widescreen {
            margin-bottom: 0.5rem;
            margin-top: 1rem;
            font-size: 1.1rem;
        }
    }

    .container {
        display: flex;
        align-items: flex-start;
        gap: 1.5rem;
        align-items: stretch;

        @include mobile {
            flex-direction: column;
        }

        @include widescreen {
            flex-direction: column;
        }

        &__export {
            background-color: var(--clr-height-1-3);
            padding: 0.9rem 1.5rem;
            border-radius: 0.2rem;

            @include widescreen {
                min-height: 7rem;
            }

            &__description {
                margin-bottom: 1rem;
                font-size: 0.8rem;

                @include widescreen {
                    font-size: 0.9rem;
                }
            }
        }
        &__import {
            background-color: var(--clr-height-1-3);
            padding: 0.9rem 1.5rem;
            border-radius: 0.2rem;

            @include widescreen {
                min-height: 7rem;
            }

            &__description {
                margin-bottom: 1rem;
                font-size: 0.8rem;

                @include widescreen {
                    font-size: 0.9rem;
                }
            }

            &__buttons {
                display: flex;
                gap: 0.9rem;

                @include mobile {
                    flex-direction: column;
                }
            }
        }
    }
</style>
