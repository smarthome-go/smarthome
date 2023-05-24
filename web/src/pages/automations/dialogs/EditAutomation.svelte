<script lang="ts">
    import Button, { Label } from '@smui/button'
    import Dialog, { Actions, Content, Header, InitialFocus, Title } from '@smui/dialog'
    import IconButton from '@smui/icon-button'
    import Switch from '@smui/switch'
    import { createEventDispatcher, onMount } from 'svelte'
    import {
        hmsLoaded,
        homescripts,
        type automation,
        type editAutomation,
        getTimeOfAutomation,
    } from '../main'
    import Inputs from './Inputs.svelte'
    import FormField from '@smui/form-field'
    import Checkbox from '@smui/checkbox'

    const days: string[] = [
        'Sunday',
        'Monday',
        'Tuesday',
        'Wednesday',
        'Thursday',
        'Friday',
        'Saturday',
    ]

    export let open = false

    // Event dispatcher
    const dispatch = createEventDispatcher()

    // Sets the previous state when the dialog is opened for the first time
    let hasUpdatedPrevious = false
    $: if (open && !hasUpdatedPrevious) {
        updatePrevious()
        hasUpdatedPrevious = true
    }

    /**
     * Data flow:
     *  - `data` is used for convenient binding from `Automation.svelte`
     *  - `inputData` is bound to the `Inputs` element
     *  - `inputDataBefore` preserves the previous state before any modification
     */

    // Bound to the `Inputs.svelte` component
    let inputData: editAutomation

    // Stores the input values before a modification
    // Is used for a rollback when using the `cancel` button
    let inputDataBefore: editAutomation

    // Only bound externally in order to use preset values
    export let data: automation

    let intervalUnit = undefined
    let intervalUnitBefore = undefined

    onMount(() => {
        const timeData =
            data.trigger === 'cron' || data.trigger === 'on_sunrise' || data.trigger === 'on_sunset'
                ? getTimeOfAutomation(data)
                : { days: [], hours: 0, minutes: 0 }

        inputData = {
            days: timeData.days,
            description: data.description,
            enabled: data.enabled,
            homescriptId: data.homescriptId,
            hour: timeData.hours,
            minute: timeData.minutes,
            name: data.name,
            trigger: data.trigger,
            disableOnce: data.disableOnce,
            triggerInterval: data.triggerInterval,
            // TODO: something is missing here
        }
        inputData['id'] = data.id
    })

    // Setting each field individually is required in order to prevent the assignment of references
    function applyCurrentState() {
        data.name = inputData.name
        data.description = inputData.description
        data.enabled = inputData.enabled
        data.homescriptId = inputData.homescriptId
        data.trigger = inputData.trigger
        data.triggerInterval = inputData.triggerInterval

        // Is used to regenerate a cron-description after modification
        let daysText = `, `
        if (inputData.days.length === 1) {
            daysText = `Only on ${days[inputData.days[0]]}`
        } else if (inputData.days.length < 7) {
            daysText = `Only on ${inputData.days
                .slice(0, inputData.days.length - 1)
                .map(d => days[d])
                .join(', ')}`
            daysText += ` and ${days[inputData.days[inputData.days.length - 1]]}`
        }
        data.cronDescription =
            `At ${inputData.hour <= 12 ? inputData.hour : inputData.hour - 12}`.padStart(2, '0') +
            ':' +
            `${inputData.minute}`.padStart(2, '0') +
            ` ${inputData.hour < 12 ? 'AM' : 'PM'} ${daysText}`
    }
    function updatePrevious() {
        inputDataBefore = {
            days: inputData.days,
            description: inputData.description,
            enabled: inputData.enabled,
            homescriptId: inputData.homescriptId,
            hour: inputData.hour,
            minute: inputData.minute,
            name: inputData.name,
            trigger: inputData.trigger,
            disableOnce: inputData.disableOnce,
            triggerInterval: inputData.triggerInterval,
        }
        inputDataBefore['id'] = data.id
        intervalUnitBefore = intervalUnit
    }
    function restorePrevious() {
        inputData = {
            days: inputDataBefore.days,
            description: inputDataBefore.description,
            enabled: inputDataBefore.enabled,
            homescriptId: inputDataBefore.homescriptId,
            hour: inputDataBefore.hour,
            minute: inputDataBefore.minute,
            name: inputDataBefore.name,
            trigger: inputDataBefore.trigger,
            disableOnce: inputDataBefore.disableOnce,
            triggerInterval: inputDataBefore.triggerInterval,
        }
        inputData['id'] = inputDataBefore["id"]
        intervalUnit = intervalUnitBefore
    }

    // Automation deletion
    let deleteOpen = false
</script>

<!-- TODO: fix before value undefined -->
{#if inputData !== undefined}
    <Dialog bind:open fullscreen aria-labelledby="title" aria-describedby="content">
        <!-- Deletion confirmation dialog -->
        <Dialog
            bind:open={deleteOpen}
            aria-labelledby="confirmation-title"
            aria-describedby="confirmation-content"
            slot="over"
        >
            <Title id="confirmation-title">Confirm Deletion</Title>
            <Content id="confirmation-content"
                >You are about to delete the automation '{data.name}'. This action will stop the
                automation from executing and remove it from the system. Are you sure you want to
                proceed?</Content
            >
            <Actions>
                <Button
                    on:click={() => {
                        dispatch('delete', null)
                    }}
                >
                    <Label>Delete</Label>
                </Button>
                <Button use={[InitialFocus]}>
                    <Label>Cancel</Label>
                </Button>
            </Actions>
        </Dialog>

        <Header>
            <Title id="title">Edit Automation</Title>
            <IconButton action="close" class="material-icons">close</IconButton>
        </Header>
        <Content id="content">
            <Inputs bind:intervalUnit bind:data={inputData} />
            <div class="actions">
                <div class="delete">
                    <Button
                        on:click={() => {
                            deleteOpen = true
                        }}
                    >
                        <Label>Delete</Label>
                    </Button>
                    <span class="text-hint"> Delete Automation </span>
                </div>
                <div class="activation">
                    <div class="activation__element">
                        <FormField>
                            <Checkbox bind:checked={inputData.disableOnce} />
                            <span class="text-hint">Disable Once</span>
                        </FormField>
                    </div>
                    <div class="activation__element">
                        <FormField>
                            <Switch bind:checked={inputData.enabled} />
                            <span class="text-hint">
                                Automation {inputData.enabled ? 'enabled' : 'disabled'}
                            </span>
                        </FormField>
                    </div>
                </div>
            </div>
        </Content>
        <Actions>
            {#if $hmsLoaded && $homescripts.length > 0}
                <Button on:click={restorePrevious}>
                    <Label>Cancel</Label>
                </Button>
                <Button
                    disabled={data.name == '' ||
                        (data.trigger === 'cron' && inputData.days.length == 0) ||
                        (data.trigger === 'interval' &&
                            (inputData.triggerInterval <= 0 ||
                                inputData.triggerInterval > 60 * 60 * 24 * 365)) ||
                        JSON.stringify(inputData) === JSON.stringify(inputDataBefore)}
                    use={[InitialFocus]}
                    on:click={() => {
                        dispatch('modify', { data: inputData, id: data.id })
                        applyCurrentState()
                        updatePrevious()
                    }}
                >
                    <Label>Edit</Label>
                </Button>
            {:else}
                <Button>
                    <Label>Cancel</Label>
                </Button>
            {/if}
        </Actions>
    </Dialog>
{/if}

<style lang="scss">
    @use '../../../mixins' as *;
    .actions {
        display: flex;
        gap: 2rem;
        align-items: center;
        background-color: var(--clr-height-0-1);
        border-radius: 0.3rem;
        padding: 1.5rem;

        div {
            width: 50%;
        }

        @include mobile {
            flex-direction: column;

            div {
                width: 100%;
            }
        }
    }

    .activation {
        display: flex;

        @include mobile {
            display: block;
        }

        &__element {
            @include mobile {
                padding: 0.5rem 0;
            }
        }
    }
</style>
