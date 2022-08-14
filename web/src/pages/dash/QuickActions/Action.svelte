<script lang="ts">
    import {
        homescriptArgSubmit,
        homescriptResponseWrapper,
        homescriptWithArgs,
        killAllJobsById,
        runHomescriptById,
    } from "../../../homescript";
    import { createSnackbar } from "../../../global";
    import HmsArgumentPrompts from "../../../components/Homescript/ArgumentPrompts/HmsArgumentPrompts.svelte";
    import ExecutionResultPopup from "../../../components/Homescript/ExecutionResultPopup/ExecutionResultPopup.svelte";
    import { createEventDispatcher } from "svelte";
    import Progress from "../../../components/Progress.svelte";
    import IconButton from "@smui/icon-button";

    // Used for dispatching events
    const dispatch = createEventDispatcher();

    export let data: homescriptWithArgs;

    // Specifies whether the current script is running
    export let running = false;

    let success = false;
    let failure = false;

    // Is used when the run button is pressed and the current script has arguments
    let argumentsPromptOpen = false;

    /*
       Executing the currently selected Homescript
    */
    let hmsExecutionResults: homescriptResponseWrapper[] = [];

    // If the current Homescript contains arguments, the function triggers the argument-prompt dialog opening
    function initCurrentRun() {
        if (data.arguments.length === 0) {
            runCurrentWithArgs([]);
            return;
        }
        // The script is executed via callback: refer to the argument dialog
        argumentsPromptOpen = true;
    }

    // Used when the run button is pressed, error handling is accomplished here
    async function runCurrentWithArgs(args: homescriptArgSubmit[]) {
        running = true;
        dispatch("run", null);
        try {
            const hmsRes = await runHomescriptById(data.data.data.id, args);

            success = hmsRes.success;
            failure = !success;

            setTimeout(() => {
                success = false;
                failure = false;
            }, 1000);

            console.log(hmsRes)

            hmsExecutionResults = [
                ...hmsExecutionResults,
                {
                    response: hmsRes,
                    code: data.data.data.code,
                    modeRun: true,
                },
            ];
        } catch (err) {
            $createSnackbar(`Could not execute ${data.data.data.name}: ${err}`);
        }
        dispatch("finish", null);
        running = false;
    }
</script>

{#if argumentsPromptOpen && data.arguments.length > 0}
    <HmsArgumentPrompts
        on:submit={(event) => {
            runCurrentWithArgs(event.detail);
        }}
        bind:open={argumentsPromptOpen}
        args={data.arguments.map((a) => a.data)}
    />
{/if}

{#if hmsExecutionResults[0] !== undefined}
    <ExecutionResultPopup
        open={true}
        data={hmsExecutionResults[0]}
        on:close={() => (hmsExecutionResults = hmsExecutionResults.slice(1))}
    />
{/if}

<div
    class="action mdc-elevation--z3"
    class:mdc-elevation--z6={running}
    on:click={initCurrentRun}
    class:running
    class:success
    class:failure
>
    <div class="action__overlay">
        <div class="action__overlay__cancel" class:hidden={!running}>
            <IconButton
                class="material-icons"
                on:click={(e) => {
                    e.stopPropagation();
                    killAllJobsById(data.data.data.id);
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
    .action {
        aspect-ratio: 1;
        height: 5rem;
        width: auto;
        flex-shrink: 1;
        border-radius: 0.25rem;
        padding: 0.5rem;
        background-color: var(--clr-height-1-3);
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        cursor: pointer;
        border: 0.1rem solid transparent;
        position: relative;
        transition-property: border-color;
        transition-duration: 0.5s;

        &.running {
            background-color: var(--clr-height-1-6);
        }

        &.success {
            border-color: var(--clr-success);
        }

        &.failure {
            border-color: var(--clr-error);
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
                &.hidden {
                    opacity: 0;
                    display: none;
                }
            }

            :global &__spinner {
                transform: scale(65%);
            }
        }

        &__icon {
            font-size: 2rem;
        }

        &__name {
            color: var(--clr-text-hint);
            font-size: 0.65rem;
            white-space: nowrap;
            text-overflow: ellipsis;
            max-width: calc(100% - 0.5rem);
            overflow: hidden;
        }
    }
</style>
