<script lang="ts">
    import Box from "../Box.svelte";
    import { createSnackbar } from "../../../global";
    import { onMount } from "svelte";
    import Reminder from "./Reminder.svelte";
    import type { reminder } from "./types";

    let reminders: reminder[] = [];
    let remindersLoaded = false;

    let loading = false;

    async function loadReminders() {
        loading = true;
        try {
            const res = (await (
                await fetch("/api/reminder/list")
            ).json()) as reminder[];
            reminders = res.sort((a, b) => b.priority - a.priority);
            remindersLoaded = true;
        } catch (err) {
            $createSnackbar("Could not load reminders");
        }
        loading = false;
    }

    onMount(loadReminders);
</script>

<Box bind:loading>
    <span slot="header">Reminders</span>
    <div
        class="reminders"
        class:empty={remindersLoaded && reminders.length === 0}
        slot="content"
    >
        {#if remindersLoaded && reminders.length === 0}
            <i class="reminders__empty__icon material-icons">done</i>
            <span class="text-hint">
                All caught up, nothing to do</span
            >
        {:else}
            {#each reminders as data (data.id)}
                <Reminder
                    bind:data
                    on:delete={() =>
                        (reminders = reminders.filter((r) => r.id !== data.id))}
                />
            {/each}
        {/if}
    </div>
</Box>

<style lang="scss">
    .reminders {
        display: flex;
        flex-direction: column;
        gap: 0.7rem;

        &.empty {
            align-items: center;
            margin-top: 2rem;
        }

        &__empty {
            &__icon {
                color: var(--clr-text-disabled);
                font-size: 7rem;
            }
        }
    }
</style>
