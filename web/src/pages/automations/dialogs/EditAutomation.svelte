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
    import { createEventDispatcher,onMount } from 'svelte'
    import {
    addAutomation,
    automation,
    generateCronExpression,
    hmsLoaded,
    homescripts,parseCronExpressionToTime
    } from '../main'
    import Inputs from './Inputs.svelte'

    export let open = false


    // Event dispatcher
    const dispatch = createEventDispatcher()


    // Binded to the `Inputs.svelte` component, will be binded to `data` reversely
    let inputsData: addAutomation

    // Only binded externally in order to use preset values
    export let data: automation

    onMount(() => {
        const timeData = parseCronExpressionToTime(data.cronExpression)
        inputsData = {
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

    let inputDataBefore = data

    function applyCurrentState() {
        data.name = inputsData.name
        data.description = inputsData.description
        data.enabled = inputsData.enabled
        data.homescriptId = inputsData.homescriptId
        data.cronDescription = generateCronExpression(
            inputsData.hour,
            inputsData.minute,
            inputsData.days
        )
        data.timingMode = inputsData.timingMode
    }
    function updatePrevious() {
        inputDataBefore = data
    }
    function restorePrevious() {
        data = inputDataBefore
        open = false
    }
</script>

{#if inputsData !== undefined}
<Dialog bind:open aria-labelledby="title" aria-describedby="content" fullscreen>
    <Header>
        <Title id="title">Edit Automation</Title>
        <IconButton action="close" class="material-icons">close</IconButton>
    </Header>
    <Content id="content">
        <Inputs bind:data={inputsData} />
    </Content>
    <Actions>
        {#if $hmsLoaded && $homescripts.length > 0}
        <Button on:click={() => (open = false)}>
            <Label>Cancel</Label>
        </Button>
        <Button
        disabled={data.name == '' || inputsData.days.length == 0}
        use={[InitialFocus]}
        on:click={() => {
           dispatch("modify", {data: inputsData, id: data.id})
           applyCurrentState()
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