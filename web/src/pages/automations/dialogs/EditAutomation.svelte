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
    import { createEventDispatcher } from 'svelte'
    import { addAutomation,automation,hmsLoaded,homescripts } from '../main'
    import Inputs from './Inputs.svelte'

    export let open = false
    // Event dispatcher
    const dispatch = createEventDispatcher()

    // Binded to the `Inputs.svelte` component
    let inputsData: addAutomation = {
        days: [],
        description: '',
        enabled: true,
        homescriptId: '',
        hour: 0,
        minute: 0,
        name: '',
        timingMode: 'normal',
    }

    export let data: automation = {
        id: 0,
	    name: "",
	    description: "",
	    cronExpression: "",
	    cronDescription: "",
	    homescriptId: "",
	    owner: "",
	    enabled: false,
	    timingMode: "normal",
    }

    // TODO: impl cron generator on change
    let inputDataBefore = data
    function updatePrevious() {
        inputDataBefore = data
    }
    function restorePrevious() {
        data = inputDataBefore
        open = false
    }
</script>

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
            <Button on:click={() => open = false}>
                <Label>Cancel</Label>
            </Button>
            <Button
                disabled={data.name == '' || inputsData.days.length == 0}
                use={[InitialFocus]}
                on:click={() => {
                    dispatch('edit', data)
                    // Reset values after creation
                    restorePrevious()
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

<style lang="scss">
</style>
