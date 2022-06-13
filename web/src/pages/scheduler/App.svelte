<script lang="ts">
    import Button, { Icon } from "@smui/button";
    import IconButton from "@smui/icon-button";
    import { Label } from "@smui/list";
    import { onMount } from "svelte";
    import Progress from "../../components/Progress.svelte";
    import { createSnackbar, data as userData } from "../../global";
    import Page from "../../Page.svelte";
    import AddSchedule from "./dialogs/AddSchedule.svelte";
    import { ScheduleData, loading, schedules } from "./main";
    import HmsEditor from "../../components/Homescript/HmsEditor/HmsEditor.svelte";

    let addOpen = false;

    // Fetches the current schedules from the server
    async function loadSchedules() {
        $loading = true;
        try {
            const res = await (
                await fetch("/api/scheduler/list/personal")
            ).json();

            if (res.success !== undefined && !res.success)
                throw Error(res.error);
        } catch (err) {
            $createSnackbar(`Could not load schedules: ${err}`);
        }
        $loading = false;
    }

    // Load the schedules as soon as possible
    onMount(loadSchedules);
</script>

<AddSchedule bind:open={addOpen} />

<Page>
    <div id="header" class="mdc-elevation--z4">
        <h6>Scheduler</h6>
        <div>
            <IconButton
                title="Refresh"
                class="material-icons"
                on:click={async () => {
                    await loadSchedules();
                }}>refresh</IconButton
            >
            {#if $schedules.length > 0}
                <Button on:click={() => (addOpen = true)}>
                    <Label>Create New</Label>
                    <Icon class="material-icons">add</Icon>
                </Button>
            {/if}
        </div>
    </div>
    <Progress id="loader" bind:loading={$loading} />

    <div class="schedules" class:empty={$schedules.length == 0}>
        {#if $schedules.length == 0}
            <i class="material-icons">event_repeat</i>
            <h6 class="text-hint">No schedules</h6>
            <Button on:click={() => (addOpen = true)} variant="outlined">
                <Label>Create New</Label>
                <Icon class="material-icons">add</Icon>
            </Button>
        {:else}
            {#each $schedules as schedule (schedule.id)}
                <span>{schedule.id}</span>
            {/each}
        {/if}
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
        h6 {
            margin: 0.5rem 0;
            @include mobile {
                // Hide title on mobile due to space limitations
                display: none;
            }
        }
    }
    .schedules {
        display: flex;
        &.empty {
            flex-direction: column;
            align-items: center;
            justify-content: center;
            padding-top: 5rem;
            color: var(--clr-text-disabled);
            gap: 1rem;
            i {
                font-size: 5rem;
            }
            h6 {
                margin: 0.5rem 0;
            }
        }
    }
</style>
