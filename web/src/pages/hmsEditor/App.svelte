<script lang="ts">
    import Terminal from "../../components/Homescript/ExecutionResultPopup/Terminal.svelte";
    import {
        homescript,
        homescriptResponseWrapper,
        homescriptWithArgs,
        lintHomescriptCode,
    } from "../../homescript";
    import { runHomescriptCode } from "../../homescript";
    import HmsEditor from "../../components/Homescript/HmsEditor/HmsEditor.svelte";
    import IconButton, { Icon } from "@smui/icon-button";
    import Progress from "../../components/Progress.svelte";
    import { onMount } from "svelte";
    import { createSnackbar } from "../../global";
    import Page from "../../Page.svelte";
    import Select, { Option } from "@smui/select";

    // Specifies whether the alternate layout (larger terminal) should be active or not
    let layoutAlt: boolean = false;

    // CTRL + S listener and default prevention
    document.addEventListener("keydown", (e) => {
        if (e.ctrlKey && e.key === "s") {
            e.preventDefault();
            saveCurrent();
        }
    });

    // Specifies whether there are unsaved changes or if the code is up-to-date
    let savedCode: string = "";

    // Is set to true when a script is linted or executed
    let requestLoading: boolean = false;
    // Is set to true if either the script loads or is saved
    let otherLoading: boolean = false;

    // Saves the last execution / lint result
    let currentExecRes: homescriptResponseWrapper = undefined;

    // Saves all available Homescripts as a list
    let homescripts: homescriptWithArgs[] = [];
    let homescriptsLoaded: boolean = false;

    $: if (homescriptsLoaded && currentScript) setCurrentScript(currentScript);

    function setCurrentScript(id: string) {
        currentData = homescripts.find(
            (h) => h.data.data.id === currentScript
        ).data;
        savedCode = currentData.data.code;
    }

    // Saves the metadata of the current script (specified by URL query)
    let currentScript: string = "";
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

    async function runCurrentCode() {
        requestLoading = true;
        try {
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
        } catch (err) {
            $createSnackbar(`Failed to run '${currentScript}': ${err}`);
        }
        requestLoading = false;
    }

    async function LintCurrentCode() {
        requestLoading = true;
        try {
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
        } catch (err) {
            $createSnackbar(`Failed to lint '${currentScript}': ${err}`);
        }
        requestLoading = false;
    }
    onMount(loadHomescript);
</script>

<Page>
    <div id="header" class="mdc-elevation--z4">
        <div id="header__left">
            <h6>Editing {currentData.data.id}</h6>
            <div
                id="header__left__save"
                class:unsaved={savedCode !== currentData.data.code}
            >
                <i class="material-icons"
                    >{savedCode === currentData.data.code
                        ? "cloud_done"
                        : "save"}</i
                >
                {savedCode === currentData.data.code ? "saved" : "unsaved"}
            </div>
        </div>
        <div id="header__buttons">
            <Select bind:value={currentScript} label="Active script">
                {#each homescripts as hms}
                    <Option value={hms.data.data.id}>{hms.data.data.id}</Option>
                {/each}
            </Select>
            <IconButton
                class="material-icons"
                on:click={() => (layoutAlt = !layoutAlt)}
                >vertical_split</IconButton
            >
            <IconButton class="material-icons" on:click={saveCurrent}
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
                <IconButton class="material-icons" on:click={runCurrentCode}
                    >play_arrow</IconButton
                >
                <IconButton class="material-icons" on:click={LintCurrentCode}>
                    bug_report</IconButton
                >
                <IconButton
                    class="material-icons"
                    on:click={() => (currentExecRes = undefined)}
                    >replay</IconButton
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
</Page>

<style lang="scss">
    @use "../../mixins" as *;

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
        }

        &__left {
            display: flex;
            gap: 1rem;

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
                overflow: auto
            }
        }
    }
</style>
