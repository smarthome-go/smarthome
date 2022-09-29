<script lang="ts">
    import Button, { Icon, Label } from "@smui/button";
    import IconButton from "@smui/icon-button";
    import type { homescript } from "../../homescript";
    import { onMount } from "svelte";
    import Progress from "../../components/Progress.svelte";
    import { createSnackbar, data as userData, hasFetched } from "../../global";
    import Page from "../../Page.svelte";
    import Automation from "./Automation.svelte";
    import AddAutomation from "./dialogs/AddAutomation.svelte";
    import Overview from "./dialogs/Overview.svelte";
    import {
        automations,
        automationsLoaded,
        hmsLoaded,
        homescripts,
        loading,
    } from "./main";
    import type { addAutomation, automation } from "./main";

    let addOpen = false;
    let overviewOpen = false;

    // Fetches the current automations from the server
    async function loadAutomations() {
        $loading = true;
        try {
            const res = await (
                await fetch("/api/automation/list/personal")
            ).json();

            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            // Group together automations which are disabled
            automations.set(
                res.sort((a: automation) => {
                    return a.enabled ? -1 : 1;
                })
            );
            $automationsLoaded = true;
        } catch (err) {
            $createSnackbar(`Could not load automations: ${err}`);
        }
        $loading = false;
    }

    // Fetches the available homescripts for the selection and naming
    async function loadHomescript() {
        $loading = true;
        try {
            let res = await (
                await fetch("/api/homescript/list/personal")
            ).json();

            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            // Filter out any homescripts which are not meant to be used for automations
            res = res.filter((a: homescript) => a.data.schedulerEnabled);
            homescripts.set(res); // Move the fetched homescripts into the store
            hmsLoaded.set(true); // Signal that the homescripts are loaded
        } catch (err) {
            $createSnackbar(`Could not load homescript: ${err}`);
        }
        $loading = false;
    }

    // Sends a request to the server to create a new automation
    async function createAutomation(data: addAutomation) {
        $loading = true;
        try {
            const res = await (
                await fetch("/api/automation/add", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify(data),
                })
            ).json();
            if (!res.success) throw Error(res.error);
            // Create a placeholder item while the automations are being updated
            // Fetching from the server is needed in order to get the generated id
            $automations = [
                ...$automations,
                {
                    cronDescription: "not yet generated",
                    cronExpression: "* * * * *",
                    description: data.description,
                    enabled: data.enabled,
                    homescriptId: data.homescriptId,
                    id: 0,
                    name: data.name,
                    owner: $userData.userData.user.username,
                    timingMode: data.timingMode,
                    disableOnce: false,
                },
            ];
            loadAutomations();
        } catch (err) {
            $createSnackbar(`Could not create automation: ${err}`);
        }
        $loading = false;
    }

    // Sends a request to the server to delete an automation
    async function deleteAutomation(id: number) {
        $loading = true;
        try {
            const res = await (
                await fetch("/api/automation/delete", {
                    method: "DELETE",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ id }),
                })
            ).json();
            if (!res.success) throw Error(res.error);
            $automations = $automations.filter((a) => a.id !== id);
        } catch (err) {
            $createSnackbar(`Could not delete automation: ${err}`);
        }
        $loading = false;
    }

    function handleAddAutomation(event) {
        const data = event.detail as addAutomation;
        createAutomation(data).then();
    }

    onMount(() => {
        loadAutomations().then(loadHomescript);
    }); // Load automations as soon as the component is mounted
</script>

<!-- Popup is shown when an automation is being added -->
<AddAutomation bind:open={addOpen} on:add={handleAddAutomation} />

<Overview bind:open={overviewOpen} />

<Page>
    <div id="header" class="mdc-elevation--z4">
        <h6>Automations</h6>
        <div id="header__content">
            {#if $userData.userData.user.schedulerEnabled}
                <IconButton
                    title="Refresh"
                    class="material-icons"
                    on:click={async () => {
                        await loadAutomations();
                        await loadHomescript();
                    }}>refresh</IconButton
                >
                {#if $automations.length > 0}
                    <IconButton
                        title="Week View"
                        class="material-icons"
                        on:click={() => (overviewOpen = true)}
                    >
                        view_list
                    </IconButton>
                    <Button on:click={() => (addOpen = true)}>
                        <Label>Create New</Label>
                        <Icon class="material-icons">add</Icon>
                    </Button>
                {/if}
            {/if}
        </div>
    </div>
    <Progress id="loader" bind:loading={$loading} />

    <div
        class="automations"
        class:empty={($automationsLoaded && $automations.length == 0) ||
            !$userData.userData.user.schedulerEnabled}
    >
        {#if !$userData.userData.user.schedulerEnabled && hasFetched}
            <div class="automations__disabled">
                <i class="material-icons" id="no-automations-icon"
                    >sync_disabled</i
                >
                <h6 class="text-hint">Automations Disabled</h6>
                <Button href="/profile" variant="outlined">
                    <Label>Enable</Label>
                </Button>
            </div>
        {:else if $automations.length == 0 && $automationsLoaded}
            <div class="automations__empty">
                <i class="material-icons" id="no-automations-icon"
                    >event_repeat</i
                >
                <h6 class="text-hint">No automations</h6>
                <Button on:click={() => (addOpen = true)} variant="outlined">
                    <Label>Create New</Label>
                    <Icon class="material-icons">add</Icon>
                </Button>
            </div>
        {:else if hasFetched && $userData.userData.user.schedulerEnabled}
            {#each $automations as automation (automation.id)}
                <Automation
                    bind:data={automation}
                    on:delete={() => deleteAutomation(automation.id)}
                    on:modify={() => {
                        // If there is at least one automation with a non-normal timing-mode, update it
                        // Fetching data from the server is required because the client does not possess information about the longitude nor the latidute of the server
                        if (
                            $automations.filter(
                                (automation) =>
                                    automation.timingMode !== "normal"
                            ).length > 0
                        ) {
                            loadAutomations();
                        }
                    }}
                />
            {/each}
        {/if}
    </div>
</Page>

<style lang="scss">
    @use "../../mixins" as *;

    .automations {
        padding: 1.5rem;
        border-radius: 0.4rem;
        display: flex;
        flex-wrap: wrap;
        gap: 1rem;
        box-sizing: border-box;

        &.empty {
            justify-content: center;
        }

        &__empty,
        &__disabled {
            justify-content: center;
            color: var(--clr-text-disabled);
            padding-top: 5rem;
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: 1rem;

            i {
                font-size: 5rem;
            }

            h6 {
                margin: 0.5rem 0;
            }
        }

        @include mobile {
            justify-content: center;
        }
    }

    #header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0.1rem 1.3rem;
        box-sizing: border-box;
        background-color: var(--clr-height-1-4);
        min-height: 3.5rem;

        &__content {
            display: flex;
            align-items: center;
        }

        h6 {
            margin: 0.5rem 0;

            @include mobile {
                // Hide title on mobile due to space limitations
                display: none;
            }
        }
    }

    #no-automations-icon {
        font-size: 5rem;
        color: var(--clr-text-disabled);
    }
</style>
