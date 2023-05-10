<script lang="ts">
    import { Icon, Label } from '@smui/button'
    import Button from '@smui/button/src/Button.svelte'
    import ConfigExport from './dialogs/ConfigExport.svelte'
    import ConfigImport from './dialogs/ConfigImport.svelte'
    import FactoryReset from './dialogs/FactoryReset.svelte'

    let exportOpen = false
    let importConfigOpen = false
    let factoryOpen = false
</script>

<ConfigExport bind:open={exportOpen} />
<ConfigImport bind:open={importConfigOpen} />
<FactoryReset bind:open={factoryOpen} />


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
        <Button on:click={() => exportOpen = true} variant="raised">
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
