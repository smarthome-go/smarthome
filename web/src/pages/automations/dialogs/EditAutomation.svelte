<script lang="ts">
    import Button,{ Label } from '@smui/button'
    import Dialog,{
    Actions,
    Content,
    Header,
    InitialFocus,
    Title
    } from '@smui/dialog'
    import IconButton from '@smui/icon-button'
    import Switch from '@smui/switch'
    import { createEventDispatcher,onMount } from 'svelte'
    import {
    addAutomation,
    automation,
    generateCronExpression,
    hmsLoaded,
    homescripts,
    parseCronExpressionToTime
    } from '../main'
    import Inputs from './Inputs.svelte'

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
     *  - `data` is used for convinient binding from `Automation.svelete`
     *  - `inputData` is binded to the `Inputs` element
     *  - `inputDataBefore` preserves the previous state before any modification
     */

    // Binded to the `Inputs.svelte` component
    let inputData: addAutomation

    // Stores the input values before a modification
    // Is used for a rollback when using the `cancel` button
    let inputDataBefore: addAutomation

    // Only binded externally in order to use preset values
    export let data: automation

    onMount(() => {
        const timeData = parseCronExpressionToTime(data.cronExpression)
        inputData = {
            days: timeData.days,
            description: data.description,
            enabled: data.enabled,
            homescriptId: data.homescriptId,
            hour: timeData.hours,
            minute: timeData.minutes,
            name: data.name,
            timingMode: data.timingMode,
        }
    })

    // Setting each field individually is required in order to prevent the assignment of references
    function applyCurrentState() {
        data.name = inputData.name
        data.description = inputData.description
        data.enabled = inputData.enabled
        data.homescriptId = inputData.homescriptId
        data.cronDescription = generateCronExpression(
            inputData.hour,
            inputData.minute,
            inputData.days
        )
        data.timingMode = inputData.timingMode
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
            timingMode: inputData.timingMode,
        }
        inputDataBefore["id"] = data.id
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
            timingMode: inputDataBefore.timingMode,
        }
    }

    // Automation deletion
    let deleteOpen = false
</script>

<!-- TODO: fix before value undefined -->
{#if inputData !== undefined}
    <Dialog
        bind:open
        fullscreen
        aria-labelledby="title"
        aria-describedby="content"
    >
        <!-- Deletion confirmation dialog -->
        <Dialog
            bind:open={deleteOpen}
            aria-labelledby="confirmation-title"
            aria-describedby="confirmation-content"
            slot="over"
        >
            <Title id="confirmation-title">Confirm Deletion</Title>
            <Content id="confirmation-content"
                >You are about to delete the automation '{data.name}'. This
                action will stop the automation from executing and remove it
                from the system. Are you shure you want to proceed?</Content
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
            <Inputs bind:data={inputData} />
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
                    <Switch bind:checked={inputData.enabled} />
                    <span class="text-hint">
                        Automation {inputData.enabled ? 'enabled' : 'disabled'}
                    </span>
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
                        inputData.days.length == 0 ||
                        JSON.stringify(inputData) ===
                            JSON.stringify(inputDataBefore)}
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
    }
</style>
