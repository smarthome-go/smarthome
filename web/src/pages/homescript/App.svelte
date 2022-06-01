<script lang="ts">
    import Button, { Icon } from "@smui/button";
    import IconButton from "@smui/icon-button";
    import { Label } from "@smui/list";
    import { onMount } from "svelte";
    import Progress from "../../components/Progress.svelte";
    import { createSnackbar, data as userData } from "../../global";
    import Page from "../../Page.svelte";
    import Inputs from "./dialogs/Inputs.svelte";
    import HmsSelector from "./dialogs/HmsSelector.svelte";
    import { hmsLoaded, homescriptData, homescripts, loading } from "./main";
import IconPicker from "src/components/IconPicker/IconPicker.svelte";

    let addOpen = false;

    let selectedDataChanged = false;
    let selection = "";
    let selectedData: homescriptData = {
        id: "",
        name: "",
        description: "",
        mdIcon: "code",
        code: "",
        quickActionsEnabled: false,
        schedulerEnabled: false,
    };

    // Using a copied `buffer` for the active script
    // Useful for a cancel feature
     $: if (selection != "") updateSelectedData();

    // Updates the `selectedDataChanged` boolean
    // Which is used to disable the action buttons conditionally

     $: if (selectedData !== undefined && selection !== "")
        updateSelectedDataChanged();

    // Depending on whether the data has changed
    function updateSelectedDataChanged() {
        const data = $homescripts.find((h) => h.data.id === selection).data;
        selectedDataChanged =
            data.name !== selectedData.name ||
            data.description !== selectedData.description ||
            data.mdIcon !== selectedData.mdIcon ||
            data.code !== selectedData.code ||
            data.schedulerEnabled !== selectedData.schedulerEnabled ||
            data.quickActionsEnabled !== selectedData.quickActionsEnabled;
    }

    // Is used as soon as the active script is changed and is not empty
    function updateSelectedData() {
        const selectedDataTemp = $homescripts.find(
            (h) => h.data.id === selection
        ).data;
        // Static, contextual data
        selectedData.id = selectedDataTemp.id; // Is required in order to send the request
        selectedData.code = selectedDataTemp.code; // Required to preserve code

        // Changeable data
        selectedData.name = selectedDataTemp.name;
        selectedData.description = selectedDataTemp.description;
        selectedData.mdIcon = selectedDataTemp.mdIcon;
        selectedData.quickActionsEnabled = selectedDataTemp.quickActionsEnabled;
        selectedData.schedulerEnabled = selectedDataTemp.schedulerEnabled;
    }

    // Is called when the changes have been successfully submitted and applied
    function updateSourceFromSelectedData() {
        // The index is required because JS clones the object
        const replaceIndex = $homescripts.findIndex(
            (h) => h.data.id === selection
        );
        $homescripts[replaceIndex].data.name = selectedData.name;
        $homescripts[replaceIndex].data.description = selectedData.description;
        $homescripts[replaceIndex].data.mdIcon = selectedData.mdIcon;
        $homescripts[replaceIndex].data.quickActionsEnabled =
            selectedData.quickActionsEnabled;
        $homescripts[replaceIndex].data.schedulerEnabled =
            selectedData.schedulerEnabled;
    }

    // Fetches the available Homescripts for the selection and naming
    async function loadHomescripts() {
        $loading = true;
        try {
            let res = await (
                await fetch("/api/homescript/list/personal")
            ).json();

            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            homescripts.set(res); // Move the fetched homescripts into the store
            hmsLoaded.set(true); // Signal that the homescripts are loaded
        } catch (err) {
            $createSnackbar(`Could not load Homescript: ${err}`);
        }
        $loading = false;
    }

    // Requests modification of the currently-selected Homescript
    async function modifyCurrentHomescript() {
        $loading = true;
        try {
            let res = await (
                await fetch("/api/homescript/modify", {
                    method: "PUT",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify(selectedData),
                })
            ).json();
            if (!res.success) throw Error(res.error);
            updateSourceFromSelectedData();
        } catch (err) {
            $createSnackbar(`Could not modify Homescript: ${err}`);
        }
        $loading = false;
    }

    onMount(() => {
        loadHomescripts();
    }); // Load homescripts as soon as the component is mounted
</script>

<Page>
    <div id="header" class="mdc-elevation--z4">
        <h6>Homescript</h6>
        <div id="header__buttons">
            <IconButton
                title="Refresh"
                class="material-icons"
                on:click={async () => {
                    await loadHomescripts();
                }}>refresh</IconButton
            >
            {#if $homescripts.length > 0}
                <Button on:click={() => (addOpen = true)}>
                    <Label>Create New</Label>
                    <Icon class="material-icons">add</Icon>
                </Button>
            {/if}
        </div>
    </div>
    <Progress id="loader" bind:loading={$loading} />

    <div id="content">
        <div id="container" class="mdc-elevation--z1">
            <div class="homescripts" class:empty={$homescripts.length == 0}>
                {#if $homescripts.length == 0}
                    <i class="material-icons" id="no-homescripts-icon"
                        >code_off</i
                    >
                    <h6 class="text-hint">No Homescripts</h6>
                    <Button
                        on:click={() => (addOpen = true)}
                        variant="outlined"
                    >
                        <Label>Create New</Label>
                        <Icon class="material-icons">add</Icon>
                    </Button>
                {:else}
                    <HmsSelector bind:selection />
                {/if}
            </div>
        </div>
        <div id="add" class="mdc-elevation--z1">
            <div class="header">
                <h6>Homescript {selection}</h6>
            </div>
            {#if $homescripts !== undefined && selection !== ""}
                <Inputs bind:data={selectedData} />

                <div class="actions">
                    <Button on:click={() => (addOpen = true)}>
                        <Label>Cancel</Label>
                    </Button>
                    <Button
                        on:click={modifyCurrentHomescript}
                        disabled={!selectedDataChanged}
                    >
                        <Label>Apply Changes</Label>
                    </Button>
                </div>
            {/if}
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

    #no-homescripts-icon {
        font-size: 5rem;
        color: var(--clr-text-disabled);
    }

    #content {
        display: flex;
        flex-direction: column;
        margin: 1rem 1.5rem;
        gap: 1rem;
        transition-property: height;
        transition-duration: 0.3s;

        @include widescreen {
            flex-direction: row;
            gap: 2rem;
        }
    }

    #container {
        background-color: var(--clr-height-0-1);
        border-radius: 0.4rem;

        @include widescreen {
            width: 50%;
        }
    }

    #add {
        background-color: var(--clr-height-0-1);
        border-radius: 0.4rem;
        padding: 1.5rem;

        h6 {
            margin: 0.5rem 0;
        }

        @include widescreen {
            width: 50%;
        }
    }

    .actions {
        display: flex;
        justify-content: flex-end;
        gap: 0.5rem;
        margin-top: 1rem;
    }
</style>
