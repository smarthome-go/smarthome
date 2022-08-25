<script lang="ts">
    import Button, { Icon } from "@smui/button";
    import IconButton from "@smui/icon-button";
    import { Label } from "@smui/list";
    import type { homescript } from "../../homescript";
    import { onMount } from "svelte";
    import Progress from "../../components/Progress.svelte";
    import { createSnackbar } from "../../global";
    import Page from "../../Page.svelte";
    import AddSchedule from "./dialogs/AddSchedule.svelte";
    import {
        hmsLoaded,
        homescripts,
        loading,
        schedules,
        switches,
        switchesLoaded,
    } from "./main";
    import type { ScheduleData } from "./main";
    import Schedule from "./Schedule.svelte";

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
            $schedules = res;
        } catch (err) {
            $createSnackbar(`Could not load schedules: ${err}`);
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

    // Loads the user's personal switches
    async function loadSwitches() {
        $loading = true;
        try {
            const res = await (await fetch("/api/switch/list/personal")).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);

            $switches = res;
            $switchesLoaded = true;
        } catch (err) {
            $createSnackbar(`Could not load switches: ${err}`);
        }
        $loading = false;
    }

    // Creates a new schedule on the server and the front end
    async function createSchedule(schedule: ScheduleData) {
        $loading = true;
        try {
            const res = await (
                await fetch("/api/scheduler/add", {
                    method: "POST",
                    headers: {
                        "Content-Type": "application/json",
                    },
                    body: JSON.stringify(schedule),
                })
            ).json();
            if (!res.success) throw Error(res.error);
            // Must refresh here in order to obtain the new id
            await loadSchedules();
        } catch (err) {
            $createSnackbar(`Could not create schedule: ${err}`);
        }
        $loading = false;
    }

    // Cancels and deletes an arbitrary schedule
    async function deleteSchedule(id: number) {
        $loading = true;
        try {
            const res = await (
                await fetch("/api/scheduler/delete", {
                    method: "DELETE",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ id }),
                })
            ).json();
            if (!res.success) throw Error(res.error);
            // Filter out the deleted schedule from the frontend
            deleteScheduleFromGUI(id);
        } catch (err) {
            $createSnackbar(`Could not cancel schedule: ${err}`);
        }
        $loading = false;
    }

    // Deletes a schedule from the frontend display but not the backend
    // Used when a schedule is executing
    function deleteScheduleFromGUI(id: number) {
        $schedules = $schedules.filter((s) => s.id !== id);
    }

    // Load the schedules as soon as possible
    onMount(async () => {
        await loadHomescript();
        await loadSwitches();
        await loadSchedules();
    });
</script>

<AddSchedule bind:open={addOpen} on:add={(e) => createSchedule(e.detail)} />

<Page>
    <div id="header" class="mdc-elevation--z4">
        <h6>Scheduler</h6>
        <div id="header__right">
            <IconButton
                title="Refresh"
                class="material-icons"
                on:click={async () => {
                    await loadHomescript();
                    await loadSwitches();
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
                <Schedule
                    bind:data={schedule}
                    on:hide={() => deleteScheduleFromGUI(schedule.id)}
                    on:delete={() => deleteSchedule(schedule.id)}
                />
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
        min-height: 3.5rem;

        h6 {
            margin: 0.5rem 0;
            @include mobile {
                // Hide title on mobile due to space limitations
                display: none;
            }
        }

        &__right {
            display: flex;
            align-items: center;
        }
    }
    .schedules {
        padding: 1.5rem;
        border-radius: 0.4rem;
        display: flex;
        flex-wrap: wrap;
        gap: 1rem;
        box-sizing: border-box;

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
