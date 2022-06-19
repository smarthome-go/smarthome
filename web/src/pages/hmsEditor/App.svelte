<script lang="ts">
    import Terminal from "../../components/Homescript/ExecutionResultPopup/Terminal.svelte";
    import {
        homescript,
        homescriptResponseWrapper,
        lintHomescriptCode,
    } from "../../homescript";
    import { runHomescriptCode } from "../../homescript";
    import HmsEditor from "../../components/Homescript/HmsEditor/HmsEditor.svelte";
    import Button from "@smui/button/src/Button.svelte";
    import { Icon, Label } from "@smui/button";
    import IconButton from "@smui/icon-button";
    import Progress from "../../components/Progress.svelte";
    import { onMount } from "svelte";

    // Is set to true when a script is linted or executed
    let requestLoading: boolean = false;
    // Is set to true if either the script loads or is saved
    let otherLoading: boolean = false;

    // Saves the metadata of the current script (specified by URL query)
    let currentScript: string = "test";
    let currentData: homescript = {
        owner: "",
        data: {
            id: currentScript,
            name: "",
            description: "",
            mdIcon: "",
            code: "",
            quickActionsEnabled: false,
            schedulerEnabled: false,
        },
    };

    async function loadCurrentData() {
        otherLoading = true;
        try {
            const res = await (
                await fetch(`/api/homescript/get/${currentScript}`)
            ).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            currentData = res;
        } catch (err) {
            // Handle the error here
            alert(err);
        }
        otherLoading = false;
    }

    async function saveCurrent() {
        otherLoading = true;
        try {
            const res = await (
                await fetch(`/api/homescript/modify`, {
                    method: "PUT",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ ...currentData.data }),
                })
            ).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
        } catch (err) {
            // TODO: add error handling
            alert(err);
        }
        otherLoading = false
    }

    async function runCurrentCode() {
        requestLoading = true;
        const currentExecResTemp = await runHomescriptCode(
            currentData.data.code,
            []
        );
        currentExecRes = {
            code: currentData.data.code,
            modeRun: true,
            response: currentExecResTemp,
        };
        if (currentData.data.code === "")
            currentExecRes.response.output = "Nothing to run.";
        requestLoading = false;
    }

    async function LintCurrentCode() {
        requestLoading = true;
        const currentExecResTemp = await lintHomescriptCode(
            currentData.data.code,
            []
        );
        currentExecRes = {
            code: currentData.data.code,
            modeRun: false,
            response: currentExecResTemp,
        };
        if (currentData.data.code === "")
            currentExecRes.response.output = "Nothing to lint.";
        requestLoading = false;
    }

    let currentExecRes: homescriptResponseWrapper = undefined;
    onMount(loadCurrentData);
</script>

<div id="header">
    <h6>Editing {currentData.data.id}</h6>
    <div id="header__buttons">
        <Progress type="circular" bind:loading={otherLoading} />
        <Button on:click={saveCurrent}>
            <Label>Save & Quit</Label>
            <Icon class="material-icons">save</Icon>
        </Button>
    </div>
</div>
<div class="container">
    <div class="container__editor">
        <HmsEditor bind:code={currentData.data.code} />
    </div>
    <div class="container__terminal">
        <div class="container__terminal__header">
            <IconButton class="material-icons" on:click={runCurrentCode}
                >play_arrow</IconButton
            >
            <IconButton class="material-icons" on:click={LintCurrentCode}>
                bug_report</IconButton
            >
            <IconButton
                class="material-icons"
                on:click={() => (currentExecRes = undefined)}>replay</IconButton
            >
        </div>
        <Progress type="linear" bind:loading={requestLoading} />
        <div class="container__terminal__content">
            {#if currentExecRes === undefined}
                <span class="gray"> This is Homescript v0.1.2 </span>
                <br />
                <span class="gray">
                    Homescript output will be displayed here.
                </span>
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
            gap: 1rem;
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
                padding: 0.5rem;
            }

            &__content {
                font-family: "Jetbrains Mono", monospace;
                font-size: 0.9rem;
                padding: 1rem;
            }
        }
    }
</style>
