<script lang="ts">
    import Button from '@smui/button'
    import IconButton from '@smui/icon-button'
    import SegmentedButton,{ Label,Segment } from '@smui/segmented-button'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'
    import HelperText from '@smui/textfield/helper-text'
    import { onMount } from 'svelte'
    import DatePicker from '../../components/DatePicker.svelte'
    import Progress from '../../components/Progress.svelte'
    import { createSnackbar } from '../../global'
    import Page from '../../Page.svelte'
    import { reminder,reminders } from './main'
    import Reminder from './Reminder.svelte'

    // Inputs for adding a reminder
    let inputName = ''
    let inputDescription = ''
    let inputDueDate = new Date()
    let selectedPriority = 'Normal'
    const priorities = ['Low', 'Normal', 'Medium', 'High', 'Urgent']

    $: console.log(inputDueDate)

    let loading = false

    async function loadReminders() {
        loading = true
        try {
            const res = (await (
                await fetch('/api/reminder/list')
            ).json()) as reminder[]
            reminders.set(res)
        } catch (err) {
            $createSnackbar('Could not load reminders')
        }
        loading = false
    }

    async function create() {
        loading = true
        try {
            const res = await (
                await fetch('/api/reminder/add', {
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        name: inputName,
                        description: inputDescription,
                        priority: priorities.indexOf(selectedPriority),
                        dueDate: inputDueDate,
                    }),
                    method: 'POST',
                })
            ).json()
            if (!res.success) throw Error()
        } catch (err) {
            $createSnackbar('Could not create reminder')
        }
        loading = false
    }

    function cancel() {
        inputName = ''
        inputDescription = ''
        selectedPriority = 'Normal'
    }

    onMount(() => loadReminders())
</script>

<Page>
    <Progress id="loader" bind:loading />
    <div id="content">
        <div id="container" class="mdc-elevation--z1">
            <div class="header">
                <h6>Reminders</h6>
                <IconButton
                    title="Refresh"
                    class="material-icons"
                    on:click={() => loadReminders()}>refresh</IconButton
                >
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
            <div id="name">
                <Textfield
                    style="width: 100%;"
                    helperLine$style="width: 100%;"
                    bind:value={inputName}
                    label="Name"
                    input$maxlength={100}
                >
                    <CharacterCounter slot="helper">0 / 100</CharacterCounter>
                </Textfield>
            </div>
            <div id="description">
                <Textfield
                    style="width: 100%;"
                    helperLine$style="width: 100%;"
                    textarea
                    bind:value={inputDescription}
                    label="Description"
                    input$rows={5}
                >
                    <HelperText slot="helper"
                        >Describe which task you want to accomplish</HelperText
                    >
                </Textfield>
            </div>
            <SegmentedButton
                segments={priorities}
                let:segment
                singleSelect
                bind:selected={selectedPriority}
            >
                <Segment {segment}>
                    <Label>{segment}</Label>
                </Segment>
            </SegmentedButton>
            <br />
            <DatePicker label={'Due Date'} bind:value={inputDueDate} />
            <br />
            <!-- Create and cancel buttons -->
            <div class="align">
                <Button
                    on:click={() => {
                        create()
                        cancel()
                    }}
                    disabled={inputName.length === 0}
                    touch
                    variant="raised"
                >
                    <Label>Create</Label>
                </Button>
                <Button
                    disabled={inputName.length === 0}
                    on:click={cancel}
                    touch
                >
                    <Label>Cancel</Label>
                </Button>
            </div>
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

    #description {
        margin-top: 1rem;
        :global(.mdc-text-field__resizer) {
            resize: none;
        }
    }

    .align {
        display: flex;
        align-items: center;
        gap: 1rem;

        @include mobile {
            gap: 0.7rem;
        }
    }
</style>
