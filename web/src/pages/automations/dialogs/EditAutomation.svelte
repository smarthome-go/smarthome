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
    import { createSnackbar } from '../../../global'
    import {
    addAutomation,
    automation,
    generateCronExpression,
    hmsLoaded,
    homescripts,
    loading,
    parseCronExpressionToTime
    } from '../main'
    import Inputs from './Inputs.svelte'

    export let open = false


    // Event dispatcher
    const dispatch = createEventDispatcher()


    // Binded to the `Inputs.svelte` component, will be binded to `data` reversely
    let inputsData: addAutomation

    // Only binded externally in order to handle reactivity
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

    async function modifyAutomation() {
        $loading = true
        try {
            inputsData['id'] = data.id
            const res = await (
                await fetch('/api/automation/modify', {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(inputsData),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            updatePrevious()
            applyCurrentState()
        } catch (err) {
            restorePrevious()
            $createSnackbar(`Could not modify automation: ${err}`)
        }
        $loading = false
    }

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
        dispatch("modify", data)
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
            modifyAutomation()
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
</style>
