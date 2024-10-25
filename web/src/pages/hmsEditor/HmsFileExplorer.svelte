<script lang="ts">
    import HmsFileOption from './HmsFileOption.svelte';
    import type { EditorHms } from './types';

    export let disabled = false
    export let homescripts: EditorHms[] = []
    export let currentScript: EditorHms = {
        unsaved: false,
        errors: false,
        data: {
            data: {
                owner: '',
                data: {
                    id: '',
                    name: '',
                    description: '',
                    mdIcon: '',
                    code: '',
                    quickActionsEnabled: false,
                    schedulerEnabled: false,
                    isWidget: false,
                    workspace: 'default',
                    type: 'NORMAL'
                },
            },
            arguments: []
        }
    }

    interface Workspace {
        name: string,
        scripts: EditorHms[],
    }

    let workspaces: Workspace[] = []
    $: updateWorkspacesFromHms(homescripts)

    function updateWorkspacesFromHms(homescripts: EditorHms[]) {
        workspaces = []
        let workspacesMap: Map<string, EditorHms[]> = new Map()

        for (let script of homescripts) {
            let newScripts = workspacesMap.get(script.data.data.data.workspace)
            if (newScripts === undefined) {
                newScripts = [script]
            } else {
                newScripts.push(script)
            }

            workspacesMap.set(
                script.data.data.data.workspace,
                newScripts,
            )
        }

        for (let [name, scripts] of workspacesMap) {
            workspaces = [...workspaces, {
                name,
                scripts,
            }]
        }
    }
</script>

<div class="explorer">
    <div class="explorer__workspaces">
        {#each workspaces as ws}
            <div class="explorer__workspaces__workspace">
                <span class="explorer__workspaces__workspace__name mdc-elevation--z8" class:disabled>
                    <i class="material-icons">folder</i>
                    <span>{ws.name}</span>
                </span>

                <div class="explorer__workspaces__workspace__scripts">
                    {#each ws.scripts as script, index}
                        <HmsFileOption
                            {disabled}
                            selected={currentScript.data.data.data.id === script.data.data.data.id}
                            endPiece={index === ws.scripts.length - 1}
                            bind:data={script}
                            on:click={() => currentScript = script}
                        ></HmsFileOption>
                    {/each}
                </div>
            </div>
        {/each}
    </div>
</div>

<style lang="scss">
    .explorer {
        width: 100%;
        height: 100%;
        background-color: var(--clr-height-0-2);

        &__workspaces {
            display: flex;
            flex-direction: column;
            justify-content: center;

            &__workspace {
                &__name {
                    display: flex;
                    align-items: center;
                    gap: .75rem;
                    padding: .3rem .7rem;
                    font-family: 'Jetbrains Mono NL', monospace;
                    font-size: .8rem;
                    line-height: 1em;
                    background-color: var(--clr-height-1-2);
                    text-overflow: elipsis;
                    overflow: hidden;

                    i {
                        font-size: .9rem;
                    }

                    &.disabled {
                        color: var(--clr-text-disabled);
                    }
                }
            }
        }
    }
</style>
