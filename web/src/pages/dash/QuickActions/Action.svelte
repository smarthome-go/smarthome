<script lang="ts">
    import { killAllJobsById, runHomescriptById } from '../../../homescript'
    import type {
        homescriptArgSubmit,
        homescriptResponseWrapper,
        homescriptWithArgs,
    } from '../../../homescript'
    import { createSnackbar } from '../../../global'
    import HmsArgumentPrompts from '../../../components/Homescript/ArgumentPrompts/HmsArgumentPrompts.svelte'
    import ExecutionResultPopup from '../../../components/Homescript/ExecutionResultPopup/ExecutionResultPopup.svelte'
    import { createEventDispatcher } from 'svelte'
    import Progress from '../../../components/Progress.svelte'
    import IconButton from '@smui/icon-button'
    import Ripple from '@smui/ripple'

    // Used for dispatching events
    const dispatch = createEventDispatcher()

    export let data: homescriptWithArgs

    // Specifies whether the current script is running
    export let running = false
    $: if (isAlreadyRunning) running = true

    export let isAlreadyRunning = false

    let success = false
    let failure = false

    // Is used when the run button is pressed and the current script has arguments
    let argumentsPromptOpen = false

    /*
       Executing the currently selected Homescript
    */
    let hmsExecutionResults: homescriptResponseWrapper[] = []

    // If the current Homescript contains arguments, the function triggers the argument-prompt dialog opening
    function initCurrentRun() {
        isAlreadyRunning = false
        if (data.arguments.length === 0) {
            runCurrentWithArgs([])
            return
        }
        // The script is executed via callback: refer to the argument dialog
        argumentsPromptOpen = true
    }

    // Used when the run button is pressed, error handling is accomplished here
    async function runCurrentWithArgs(args: homescriptArgSubmit[]) {
        running = true
        dispatch('run', null)
        try {
            const hmsRes = await runHomescriptById(data.data.data.id, args, false)

            success = hmsRes.success
            failure = !success

            // TODO: what is the purpose of this timeout?
            setTimeout(() => {
                success = false
                failure = false
            }, 2000)

            hmsExecutionResults = [
                ...hmsExecutionResults,
                {
                    modeRun: true,
                    response: {
                        title: data.data.data.name,
                        success: hmsRes.success,
                        output: hmsRes.output,
                        fileContents: new Map(),
                        errors: hmsRes.errors,
                    },
                },
            ]
        } catch (err) {
            $createSnackbar(`Could not execute ${data.data.data.name}: ${err}`)
        }
        dispatch('finish', null)
        running = false
    }
</script>

{#if argumentsPromptOpen && data.arguments.length > 0}
    <HmsArgumentPrompts
        on:submit={event => {
            runCurrentWithArgs(event.detail)
        }}
        bind:open={argumentsPromptOpen}
        args={data.arguments.map(a => a.data)}
    />
{/if}

{#if hmsExecutionResults[0] !== undefined}

                        <!-- bind:open={errorsOpen} -->
                        <!-- data={{ -->
                        <!--     modeRun: true, -->
                        <!--     response: { -->
                        <!--     }, -->
                        <!-- }} -->
                        <!-- on:close={() => { -->
                        <!--     // This hack is required so that the window still remains scrollable after removal -->
                        <!-- }} -->

    <ExecutionResultPopup
        open={hmsExecutionResults[0] !== undefined}
        data={hmsExecutionResults[0]}
        on:close={() => {
            // This hack is required so that the window still remains scrollable after removal
            setTimeout(() => (hmsExecutionResults = hmsExecutionResults.slice(1)), 1000)
        }}
    />
{/if}

<div
    class="action mdc-elevation--z4"
    class:running
    class:success
    class:failure
    use:Ripple={{ surface: !running }}
    on:keydown={running ? () => $createSnackbar('This action is already running') : initCurrentRun}
    on:click={running ? () => $createSnackbar('This action is already running') : initCurrentRun}
>
    <div class="action__overlay">
        <div class="action__overlay__cancel" class:hidden={!running}>
            <IconButton
                class="material-icons"
                on:click={e => {
                    e.stopPropagation()
                    killAllJobsById(data.data.data.id)
                    if (isAlreadyRunning) {
                        running = false
                        dispatch('finish', null)
                    }
                }}
                size="button"
            >
                cancel
            </IconButton>
        </div>
        <div class="action__overlay__spinner">
            <Progress bind:loading={running} type="circular" />
        </div>
    </div>
    <i class="action__icon material-icons">{data.data.data.mdIcon}</i>
    <span class="action__name">
        {data.data.data.name}
    </span>
</div>

<style lang="scss">
    @use '../../../_mixins.scss' as *;

    .action {
        aspect-ratio: 1;
        height: auto;
        width: 4.5rem;
        max-width: 5rem;
        flex-shrink: 1;
        border-radius: 0.125rem;
        padding: 0.5rem;
        background-color: var(--clr-height-3-4);
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        cursor: pointer;
        position: relative;
        user-select: none;

        // Account for the success / failure color transition
        transition-property: color;
        transition-duration: 1s;

        @include mobile {
            width: 3.9rem;
        }

        &.running {
            opacity: 60%;
            cursor: default;

            .action__icon,
            .action__name {
                transform: translateY(0.6rem);
            }
        }

        &.success {
            color: var(--clr-success);
        }

        &.failure {
            color: var(--clr-error);
        }

        &__overlay {
            width: 100%;
            position: absolute;
            top: 0;
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 0.2rem;
            box-sizing: border-box;

            &__cancel {
                color: var(--clr-error);
                z-index: 100;

                &.hidden {
                    display: none;
                }
            }

            :global &__spinner {
                transform: scale(65%);
            }
        }

        &__icon {
            font-size: 1.7rem;
            color: var(--clr-text-hint);
            transition-property: transform;
            transition-duration: 0.1s;
        }

        &__name {
            color: var(--clr-text-hint);
            font-size: 0.65rem;
            white-space: nowrap;
            text-overflow: ellipsis;
            max-width: calc(100% - 0.5rem);
            overflow: hidden;
            transition-property: transform;
            transition-duration: 0.1s;
        }
    }
</style>
