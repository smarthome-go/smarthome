<script lang="ts">
    import Page from "../../Page.svelte";
    import { createSnackbar } from "../../global";
    import {
        killAllJobsById,
        lintHomescriptCode,
        runHomescriptById,
    } from "../../homescript";
    import type {
        homescript,
        homescriptArgSubmit,
        homescriptResponseWrapper,
        homescriptWithArgs,
    } from "../../homescript";
    import { onMount } from "svelte";
    import IconButton from "@smui/icon-button";
    import Select, { Option } from "@smui/select";
    import Progress from "../../components/Progress.svelte";

    // Custom HMS components
    import HmsEditor from "../../components/Homescript/HmsEditor/HmsEditor.svelte";
    import Terminal from "../../components/Homescript/ExecutionResultPopup/Terminal.svelte";
    import Button, { Label } from "@smui/button";
    import HmsArgumentPrompts from "../../components/Homescript/ArgumentPrompts/HmsArgumentPrompts.svelte";

    /*
       General variables
       Includes varialbes such as layout-management and loading indicators
     */
    // Specifies whether the argument prompt dialog should be open or closed
    let argumentsPromptOpen = false;

    // Specifies whether the alternate layout (larger terminal) should be active or not
    let layoutAlt = false;

    // Is set to true when a script is linted or executed
    let requestLoading = false;

    // Is set to true if either the script loads or is saved
    let otherLoading = false;

    // If set to true, a banner (indicating that no script xyz has been found) is shown instead of the editor
    let err404 = false;

    // Specifies the amount of jobs executing the current id (fetched initially)
    let currentExecutionCount = 0;

    // Specifies the amount of jobs which the browser currently waits for
    // Used to limit the number of concurrent operations to exactly 1
    let currentExecutionHandles = 0;

    /*
       Script management
       Variables and functions required to save all scripts
     */
    let homescripts: homescriptWithArgs[] = [];

    // Is set to true as soon as the scripts are loaded
    // Required in the dynamic update of the current script (due to the list being empty when loaded=false)
    let homescriptsLoaded = false;

    async function loadHomescript() {
        otherLoading = true;
        try {
            const res = await (
                await fetch(`/api/homescript/list/personal/complete`)
            ).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            homescripts = res;
            homescriptsLoaded = true;
            if (homescripts.length > 0)
                currentScript = homescripts[0].data.data.id;
        } catch (err) {
            $createSnackbar(
                `Failed to load editor for '${currentScript}': ${err}`
            );
        }
        otherLoading = false;
    }

    /*
       Current script management
       Saves which script is currently being edited
       Includes a function for changing the currently active script
     */
    // Saves the metadata of the current script (specified by URL query)
    let currentScript = "";

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
            workspace: "default",
        },
    };

    // Is called every time the `currentScript` variable changes
    $: if (homescriptsLoaded && currentScript) setCurrentScript();

    // Is used to update the currently shown script
    function setCurrentScript() {
        currentData = homescripts.find(
            (h) => h.data.data.id === currentScript
        ).data;
        savedCode = currentData.data.code;
    }

    /*
       Code saving
       Variables and functions responsible for saving the code which is currently being edited
    */
    // Specifies whether there are unsaved changes or if the code is up-to-date
    let savedCode = "";

    // KEY BINDS
    // CTRL + S => Save current script
    // F8       => Run current script
    // F9       => Lint current code
    // F10      => Cancel current job(s)
    document.addEventListener("keydown", (e) => {
        if (e.ctrlKey && e.key === "s") {
            e.preventDefault();
            saveCurrent();
        } else if (e.key === "F8") {
            if (currentExecutionHandles > 0) return;
            e.preventDefault();
            initCurrentRun();
        } else if (e.key === "F9") {
            if (currentExecutionHandles > 0) return;
            e.preventDefault();
            initCurrentLint();
        } else if (e.key === "F10") {
            e.preventDefault();
            killAllJobsOfCurrent();
        }
    });

    // Sends a `save` request to the server, also updates the GUI display of unsaved changes to saved
    async function saveCurrent() {
        if (savedCode === currentData.data.code) return;
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
            savedCode = currentData.data.code;
        } catch (err) {
            $createSnackbar(`Failed to save '${currentScript}': ${err}`);
        }
        otherLoading = false;
    }

    /*
       Execution and linting
       Functions and variables responsible for linting and running the code
     */
    // Saves the last execution / lint result
    let currentExecRes: homescriptResponseWrapper = undefined;

    // Keeps track of whether the current HMS request is meant to be `run` or `lint`
    // Is used in the argument-prompt popup which dispatches the according request to the server
    let currentExecModeLint = false;

    // If the current Homescript contains arguments, the function triggers the argument-prompt dialog opening
    // Ported from `src/pages/homescript/App.svelte`
    function initCurrentRun() {
        if (
            homescripts.find((h) => h.data.data.id === currentScript).arguments
                .length === 0
        ) {
            runCurrentCode([]);
            return;
        }
        // The script is executed via callback: refer to the argument dialog
        currentExecModeLint = false;
        argumentsPromptOpen = true;
    }

    // Normal run functions
    async function runCurrentCode(args: homescriptArgSubmit[]) {
        requestLoading = true;
        currentExecutionCount++;
        currentExecutionHandles++;
        try {
            const currentExecResTemp = await runHomescriptById(
                currentData.data.id,
                args
            );
            currentExecRes = {
                code: currentData.data.code,
                modeRun: true,
                response: currentExecResTemp,
            };
            if (currentData.data.code === "")
                currentExecRes.response.output = "Nothing to run.";
        } catch (err) {
            $createSnackbar(`Failed to run '${currentScript}': ${err}`);
        }
        currentExecutionCount--;
        currentExecutionHandles--;
        requestLoading = false;
    }

    // If the current Homescript contains arguments, the function triggers the argument-prompt dialog opening
    // Ported from `src/pages/homescript/App.svelte`
    function initCurrentLint() {
        if (
            homescripts.find((h) => h.data.data.id === currentScript).arguments
                .length === 0
        ) {
            lintCurrentCode([]);
            return;
        }
        // The script is linted via callback: refer to the argument dialog
        currentExecModeLint = true;
        argumentsPromptOpen = true;
    }

    // Dry-run function without data modifications or expensive operations
    // Can be used to validate the correctness of a script without the need for execution
    async function lintCurrentCode(args: homescriptArgSubmit[]) {
        requestLoading = true;
        currentExecutionCount++;
        currentExecutionHandles++;
        try {
            const currentExecResTemp = await lintHomescriptCode(
                currentData.data.code,
                args
            );
            currentExecRes = {
                code: currentData.data.code,
                modeRun: false,
                response: currentExecResTemp,
            };
            if (currentData.data.code === "")
                currentExecRes.response.output = "Nothing to lint.";
        } catch (err) {
            $createSnackbar(`Failed to lint '${currentScript}': ${err}`);
        }
        currentExecutionCount--;
        currentExecutionHandles--;
        requestLoading = false;
    }

    // Kill all HMS jobs which are running the current Homescript-Id
    async function killAllJobsOfCurrent() {
        requestLoading = true;
        try {
            await killAllJobsById(currentData.data.id);
        } catch (err) {
            $createSnackbar(`Failed to lint '${currentScript}': ${err}`);
        }
        requestLoading = false;
    }

    // Load the Homescript-list at the beginning
    onMount(async () => {
        await loadHomescript();
        // Used for initially setting the active script via URL query
        const selectedFromQuery = new URLSearchParams(
            window.location.search
        ).get("id");
        if (
            homescripts.find((h) => h.data.data.id === selectedFromQuery) ===
            undefined
        ) {
            err404 = true;
            return;
        }
        currentScript = selectedFromQuery;
    });
</script>

{#if argumentsPromptOpen && homescripts.find((h) => h.data.data.id === currentScript).arguments.length > 0}
    <HmsArgumentPrompts
        on:submit={(event) => {
            // Handle the decision between lint and run here
            if (currentExecModeLint) {
                lintCurrentCode(event.detail);
            } else runCurrentCode(event.detail);
        }}
        bind:open={argumentsPromptOpen}
        args={homescripts
            .find((h) => h.data.data.id === currentScript)
            .arguments.map((a) => a.data)}
    />
{/if}

<Page>
    {#if err404}
        <div id="error404">
            <i class="material-icons" id="no-automations-icon">edit_off</i>
            <h6 class="text-hint">Not found</h6>
            <Button on:click={() => (window.location.href = "/homescript")}>
                <Label>View Homescripts</Label>
            </Button>
        </div>
    {:else}
        <div id="header" class="mdc-elevation--z4">
            <div id="header__left">
                <span>Editing {currentData.data.name} </span>
                <div
                    id="header__left__save"
                    class:unsaved={savedCode !== currentData.data.code}
                >
                    <i class="material-icons"
                        >{savedCode === currentData.data.code
                            ? "cloud_done"
                            : "backup"}</i
                    >
                    {savedCode === currentData.data.code ? "saved" : "unsaved"}
                </div>
            </div>
            <div id="header__buttons">
                <Select
                    bind:value={currentScript}
                    label="Preview script"
                    disabled={currentExecutionHandles !== 0}
                >
                    {#each homescripts as hms}
                        <Option value={hms.data.data.id}
                            >{hms.data.data.id}</Option
                        >
                    {/each}
                </Select>
                <IconButton
                    class="material-icons"
                    on:click={() => (layoutAlt = !layoutAlt)}
                    >vertical_split</IconButton
                >
                <IconButton
                    class="material-icons"
                    on:click={saveCurrent}
                    disabled={savedCode === currentData.data.code}
                    >save</IconButton
                >
                <Progress type="circular" bind:loading={otherLoading} />
            </div>
        </div>
        <div class="container">
            <div class="container__editor" class:alt={layoutAlt}>
                <HmsEditor bind:code={currentData.data.code} />
            </div>
            <div class="container__terminal" class:alt={layoutAlt}>
                <div class="container__terminal__header mdc-elevation--z2">
                    <IconButton
                        class="material-icons"
                        on:click={initCurrentRun}
                        disabled={savedCode !== currentData.data.code ||
                            currentExecutionHandles > 0}>play_arrow</IconButton
                    >
                    <IconButton
                        class="material-icons"
                        on:click={initCurrentLint}
                        disabled={currentExecutionHandles > 0}
                    >
                        bug_report</IconButton
                    >
                    <IconButton
                        class="material-icons"
                        on:click={killAllJobsOfCurrent}
                        disabled={currentExecutionCount === 0}
                    >
                        cancel</IconButton
                    >
                    <IconButton
                        class="material-icons"
                        on:click={() => (currentExecRes = undefined)}
                        disabled={currentExecutionHandles > 0}
                        >replay</IconButton
                    >
                </div>
                <Progress type="linear" bind:loading={requestLoading} />
                <div class="container__terminal__content">
                    {#if currentExecutionHandles > 0}
                        <span class="gray">
                            Homescript is executing: Waiting for server...
                        </span>
                    {:else if currentExecRes === undefined}
                        <span class="gray">
                            Homescript output will be displayed here.
                        </span>
                    {:else}
                        <Terminal data={currentExecRes} />
                    {/if}
                </div>
            </div>
        </div>
    {/if}
</Page>

<style lang="scss">
    @use "../../mixins" as *;

    #error404 {
        display: flex;
        flex-direction: column;
        align-items: center;
        margin-top: 8.5rem;
        gap: 1rem;

        i {
            font-size: 5rem;
            color: var(--clr-text-disabled);
        }

        h6 {
            margin: 0.5rem 0;
        }
    }

    #header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0.1rem 1.3rem;
        box-sizing: border-box;
        background-color: var(--clr-height-1-4);
        height: 3.5rem;

        @include mobile {
            // Hides bottom line
            height: 3.4rem;
            justify-content: flex-start;
        }

        &__left {
            display: flex;
            gap: 1rem;

            span {
                @include mobile {
                    display: none;
                }
            }

            &__save {
                color: var(--clr-text-disabled);
                display: flex;
                align-items: center;
                gap: 0.4rem;
                font-size: 0.9rem;

                @include mobile {
                    display: none;
                }

                i {
                    font-size: 1.25em;
                }

                &.unsaved {
                    color: var(--clr-error);
                }
            }
        }

        &__buttons {
            display: flex;
            align-items: center;
            gap: 0.2rem;
        }
    }

    .container {
        display: flex;
        overflow: hidden;
        flex-direction: column;
        height: calc(100vh - 3.67rem);

        @include mobile {
            height: calc(100vh - 9rem);
        }

        @include widescreen {
            flex-direction: row;
        }

        &__editor {
            overflow: auto;
            height: 75%;

            @include widescreen {
                width: 75%;
                height: 100%;
            }

            // Used when the expand-terminal button is selected
            transition-property: width, height;
            transition-duration: 0.25s;

            &.alt {
                @include widescreen {
                    width: 25%;
                }
                @include not-widescreen {
                    height: 25%;
                }
            }
        }

        &__terminal {
            height: 25%;

            // Used when the expand-terminal button is selected
            transition-property: width, height;
            transition-duration: 0.25s;

            @include widescreen {
                width: 25%;
                height: 100%;
            }

            &.alt {
                @include widescreen {
                    width: 75%;
                }
                @include not-widescreen {
                    height: 75%;
                }
            }

            &__header {
                padding: 0.2rem;
                background-color: var(--clr-height-0-1);
            }

            &__content {
                font-family: "Jetbrains Mono", monospace;
                font-size: 0.9rem;
                padding: 1rem;
                height: 100%;
                overflow: auto;
            }
        }
    }
</style>
