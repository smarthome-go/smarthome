<script lang="ts">
    import IconButton from '@smui/icon-button'
    import { onMount } from 'svelte'
    import Progress from '../../components/Progress.svelte'
    import { createSnackbar } from '../../global'
    import Page from '../../Page.svelte'
    import Inputs from './Inputs.svelte'
    import { loading,reminder,reminders } from './main'
    import Reminder from './Reminder.svelte'

    $loading = false
  
    async function loadReminders() {
        $loading = true
        try {
            const res = (await (
                await fetch('/api/reminder/list')
            ).json()) as reminder[]
            reminders.set(res)
        } catch (err) {
            $createSnackbar('Could not load reminders')
        }
        $loading = false
    }

      // Create mode
      async function create(name, description, priority, dueDate) {
        $loading = true
        try {
            const res = await (
                await fetch('/api/reminder/add', {
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        name,
                        description,
                        priority,
                        dueDate: dueDate.getTime(),
                    }),
                    method: 'POST',
                })
            ).json()
            if (!res.success) throw Error(`request error: ${res.error}`)
            await loadReminders()
        } catch (err) {
            $createSnackbar(`Could not create reminder ${err}`)
        }
        $loading = false
    }
    
    onMount(() => loadReminders())
</script>

<Page>
    <Progress id="loader" bind:loading={$loading} />
    <div id="content">
        <div id="container" class="mdc-elevation--z1">
            <div class="header">
                <h6>Add Reminder</h6>
                <!-- <IconButton
                    title="Refresh"
                    class="material-icons"
                    on:click={() => loadReminders()}>refresh</IconButton
                > -->
            </div>
            <div class="reminders" class:empty={$reminders.length === 0}>
                {#if $reminders.length === 0}
                    No reminders
                {/if}
                {#each $reminders as reminder (reminder.id)}
                    <Reminder {...reminder} />
                {/each}
            </div>
        </div>
        <div id="add" class="mdc-elevation--z1">
            <div class="header">
                <h6>Reminders</h6>
                <IconButton
                    title="Refresh"
                    class="material-icons"
                    on:click={() => loadReminders()}>refresh</IconButton
                >
            </div>
            <Inputs     onSubmit={create} submitLabel={"create"}/>  
        </div>
    </div>
</Page>

<style lang="scss">
    @use '../../mixins' as *;

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
        padding: 1.5rem;

        @include widescreen {
            width: 50%;
        }
    }

    #add {
        background-color: var(--clr-height-0-1);
        border-radius: 0.4rem;
        padding: 1.5rem;

        @include widescreen {
            width: 50%;
        }
    }

    .reminders {
        padding: 1rem 0;
        display: flex;
        flex-direction: column;
        overflow-x: hidden;

        &.empty {
            display: flex;
            align-items: center;
            justify-content: center;
        }
    }

    .header {
        display: flex;
        justify-content: space-between;

        h6 {
            margin: 0;
        }
    }
</style>
