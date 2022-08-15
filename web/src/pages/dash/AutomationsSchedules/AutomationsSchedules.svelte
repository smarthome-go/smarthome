<script lang="ts">
    import Box from "../Box.svelte";
    import { createSnackbar } from "../../../global";
    import { onMount } from "svelte";

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

    let schedules: Schedule[] = [];
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
            schedulesLoaded = true
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
    <div slot="content">
        {automations.length} automations registered.
        <br />
        Disabled {automations.filter((a) => !a.enabled).length}
        <br>
        <br>
        {schedules.length} schedules registered.
        <br />
    </div>
</Box>
