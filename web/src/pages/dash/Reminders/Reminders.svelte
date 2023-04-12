<script lang="ts">
    import Box from '../Box.svelte'
    import { createSnackbar } from '../../../global'
    import { onMount } from 'svelte'
    import Reminder from './Reminder.svelte'
    import type { reminder } from './types'

    let reminders: reminder[] = []
    let remindersLoaded = false

    let loading = false

    async function loadReminders() {
        loading = true
        try {
            const res = (await (await fetch('/api/reminder/list')).json()) as reminder[]
            sortReminders(res)
            remindersLoaded = true
        } catch (err) {
            $createSnackbar('Could not load reminders')
        }
        loading = false
    }

    function sortReminders(input: reminder[]) {
        reminders = input.sort((a, b) => {
            // Sort by priority
            if (b.priority !== a.priority) {
                return b.priority - a.priority
            }
            // then sort by due date
            return a.dueDate - b.dueDate
        })
    }

    function deleteReminder(id: number) {
        sortReminders(reminders.filter(r => r.id !== id))
    }

    onMount(loadReminders)
</script>

<Box bind:loading>
    <a href="/reminders" slot="header" class="title">Reminders</a>
    <div class="reminders" class:empty={remindersLoaded && reminders.length === 0} slot="content">
        {#if remindersLoaded && reminders.length === 0}
            <i class="reminders__empty__icon material-icons">done</i>
            <span class="text-hint"> All caught up, nothing to do</span>
        {:else}
            {#each reminders as data (data.id)}
                <Reminder bind:data on:delete={() => deleteReminder(data.id)} />
            {/each}
        {/if}
    </div>
</Box>

<style lang="scss">
    .title {
        color: var(--clr-primary-hint);
        font-weight: bold;
        text-decoration: none;
    }
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
