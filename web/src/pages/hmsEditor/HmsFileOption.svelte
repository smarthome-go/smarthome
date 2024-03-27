<script lang="ts">
    import type { EditorHms } from './types'
    import Ripple from '@smui/ripple';
    import { createEventDispatcher } from 'svelte';

    let dispatch = createEventDispatcher()

    export let selected: boolean = false
    export let endPiece: boolean = false

    $: console.log(data.unsaved)

    export let data: EditorHms = {
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
            arguments: [],
        }
    }

    function onClick() {
        dispatch('click')
    }
</script>

<div
    class="fileOption"
    class:selected
    class:unsaved={data.unsaved}
    class:errors={data.errors}
    use:Ripple={{ surface: true }}
    tabindex="0"
    role="button"
    on:click={onClick}
    on:keypress={onClick}
>
    <div class="fileOption__border-left" class:endPiece></div>
    <div class="fileOption__border-tpiece" class:endPiece></div>

    <i class="material-icons">
        {#if data.data.data.data.type === 'DRIVER'}
           memory
        {:else}
            {data.data.data.data.mdIcon}
        {/if}
    </i>

    <div class="fileOption__right">
        <span class="fileOption__right__id">{data.data.data.data.id}</span>
        <span class="text-hint fileOption__right__name">{data.data.data.data.name}</span>
    </div>

    <div class="fileOption__unsaved-indicator"></div>
</div>

<style lang="scss">
    .fileOption {
        display: flex;
        align-items: stretch;
        $outlineClr: var(--clr-text-disabled);
        padding-right: .5rem;

        * {
            user-select: none;
        }

        &__border-left {
            margin-left: 1rem;
            border-left: dashed .15rem $outlineClr;

            &.endPiece {
                transform: translateY(-50%);
            }
        }

        &__border-tpiece {
            border-bottom: dashed .15rem $outlineClr;
            width: 100%;
            max-width: 1.5rem;
            transform: translateY(-1rem);
        }

        i {
            font-size: 1rem;
            color: var(--clr-text-hint);
            display: flex;
            align-items: center;
            padding-left: .5rem;
        }

        &__right {
            display: flex;
            justify-content: center;
            flex-direction: column;
            gap: .3rem;
            padding: .3rem 0;
            padding-left: .5rem;

            &__id {
                font-family: "Jetbrains Mono", monospace;
                font-size: .75rem;
                line-height: 1em;
            }

            &__name {
                font-size: .6rem;
                line-height: 1em;
            }
        }

        &.errors {
            .fileOption__right > * {
                color: var(--clr-error);
            }
        }

        &.selected {
            background-color: var(--clr-height-2-6);
        }

        &.unsaved {
            .fileOption__unsaved-indicator {
                align-self: center;

                $size: .5rem;
                height: $size;
                width: $size;

                background-color: var(--clr-text-hint);
                border-radius: 50%;
                margin-left: auto;
            }
        }
    }
</style>
