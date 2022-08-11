<script lang="ts">
    import IconButton from "@smui/icon-button/src/IconButton.svelte";
    import Select, { Option } from "@smui/select";
    import { onMount } from "svelte";

    import Progress from "../../../components/Progress.svelte";
    import { createSnackbar } from "../../../global";
    import { levels, logs } from "../main";
    import LogEvent from "./LogEvent.svelte";

    // Specifies whether the loading indicator in the logs list should be active or not
    let loading = false;

    let minLevel = "TRACE";

    async function fetchLogs() {
        loading = true;
        try {
            const res = await (await fetch("/api/logs/list/all")).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            $logs = res;
        } catch (err) {
            $createSnackbar(`Failed to load system event logs: ${err}`);
        }
        loading = false;
    }

    // As soon as the component is mounted, fetch the logs
    onMount(fetchLogs);
</script>

<div class="logs">
    <div class="logs__header">
        <h6>Event Logs</h6>
        <Progress type="linear" bind:loading />
        <IconButton class="material-icons">delete</IconButton>
        <IconButton class="material-icons">expand_more</IconButton>
        <IconButton class="material-icons">expand_less</IconButton>

        <Select bind:value={minLevel} label="Minimul Level">
            {#each levels as level}
                <Option value={level}>
                    <span
                        style:color={level.color}>
                        {level.label}
                    </span>
                </Option>
            {/each}
        </Select>
    </div>
    <div class="logs__list">
        {#each $logs.filter((e) => e.level >= levels.findIndex((l) => l.label === minLevel)) as event (event.id)}
            <LogEvent data={event} />
        {/each}
    </div>
</div>

<style lang="scss">
    .logs {
        &__header {
            padding: 0.5rem 1rem;
            h6 {
                margin: 0.5rem 0;
            }
        }

        &__list {
            display: flex;
            flex-direction: column;
            gap: 1rem;
            padding: 0 1rem;
            overflow-y: auto;
        }
    }
</style>
