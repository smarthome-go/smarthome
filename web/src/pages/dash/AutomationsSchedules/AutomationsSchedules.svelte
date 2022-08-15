<script lang="ts">
    import Box from "../Box.svelte";
    import { createSnackbar } from "../../../global";
    import { onMount } from "svelte";
    import type { automation, Schedule as ScheduleType } from "./types";
    import Schedule from "./Schedule.svelte";
    import Automation from "./Automation.svelte";

    let loading = false;

    let automations: automation[] = [];
    let automationsLoaded = false;

    // Fetches the current automations from the server
    async function loadAutomations() {
        loading = true;
        try {
            const res = await (
                await fetch("/api/automation/list/personal")
            ).json();

            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            automations = res;
            automationsLoaded = true;
        } catch (err) {
            $createSnackbar(`Could not load automations: ${err}`);
        }
        loading = false;
    }

    let schedules: ScheduleType[] = [];
    let schedulesLoaded = false;

    // Fetches the current schedules from the server
    async function loadSchedules() {
        loading = true;
        try {
            const res = await (
                await fetch("/api/scheduler/list/personal")
            ).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            schedules = res;
            schedulesLoaded = true;
        } catch (err) {
            $createSnackbar(`Could not load schedules: ${err}`);
        }
        loading = false;
    }

    onMount(async () => {
        await loadAutomations();
        await loadSchedules();
    });
</script>

<Box bind:loading>
    <span slot="header">Schedules and Automations</span>
    <div class="content" slot="content">
        <div class="content__automations">
            {automations.length} automation{automations.length !== 1 ? "s" : ""}
            registered. Disabled {automations.filter((a) => !a.enabled).length}
            <div class="content__automnations__list">
                {#each automations.filter((a) => !a.enabled) as data}
                    <Automation bind:data />
                {/each}
            </div>
        </div>
        <div class="content__schedules">
            <span class="content__schedules__title">
                {schedules.length} schedule{schedules.length !== 1 ? "s" : ""} registered.
            </span>

            <div class="content__schedules__list">
                {#each schedules as data}
                    <Schedule bind:data />
                {/each}
            </div>
        </div>
    </div>
</Box>

<style lang="scss">
    .content {
        display: flex;
        gap: 1rem;

        &__automations {
            width: 50%;
        }

        &__schedules {
            width: 50%;

            &__title {
                color: var(--clr-text-hint);
            }

            &__list {
                display: flex;
                flex-direction: column;
            }
        }
    }
</style>
