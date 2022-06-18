<script lang="ts">
    import Terminal from "../../components/Homescript/ExecutionResultPopup/Terminal.svelte";
    import type { homescriptResponseWrapper } from "../../homescript";
    import { runHomescriptCode } from "../../homescript";
    import HmsEditor from "../../components/Homescript/HmsEditor/HmsEditor.svelte";
    import Button from "@smui/button/src/Button.svelte";
    import { Icon, Label } from "@smui/button";
    import IconButton from "@smui/icon-button";

    let loading: boolean = false;

    async function runCurrentCode() {
        loading = true;
        const currentExecResTemp = await runHomescriptCode(code, []);
        currentExecRes = {
            code: code,
            modeLint: false,
            response: currentExecResTemp
        };
        loading = false;
    }

    let code: string = "";

    let currentExecRes: homescriptResponseWrapper = undefined;
</script>

<div id="header">
    <h6>Editing xyz</h6>
    <div id="header__buttons">
        <Button>
            <Label>Save & Quit</Label>
            <Icon class="material-icons">save</Icon>
        </Button>
    </div>
</div>
<div class="container">
    <div class="container__editor">
        <HmsEditor bind:code/>
    </div>
    <div class="container__terminal">
        <div class="container__terminal__header">
            <IconButton class="material-icons" on:click={runCurrentCode}>play_arrow</IconButton>
            <IconButton class="material-icons">bug_report</IconButton>
            <IconButton class="material-icons" on:click={() => currentExecRes = undefined}>replay</IconButton>
        </div>
        <div class="container__terminal__content">
            {#if currentExecRes === undefined}
                START SCRIPT FIRST
            {:else}
                <Terminal data={currentExecRes} />
            {/if}
        </div>
    </div>
</div>

<style lang="scss">
    @use "../../mixins" as *;

    #header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0.1rem 1.3rem;
        box-sizing: border-box;
        background-color: var(--clr-height-1-4);
        min-height: 3.5rem;

        &__buttons {
            display: flex;
            align-items: center;
        }

        h6 {
            margin: 0.5em 0;

            @include mobile {
                // Hide title on mobile due to space limitations
                display: none;
            }
        }
    }

    .container {
        display: flex;
        overflow: hidden;
        height: calc(100vh - 3.5rem);

        &__editor {
            width: 75%;
            overflow: auto;
        }

        &__terminal {
            width: 25%;

            &__header {
                background-color: var(--clr-height-0-1);
            }

            &__content {
                font-family: "Jetbrains Mono", monospace;
                padding: 1rem;
            }
        }
    }
</style>
