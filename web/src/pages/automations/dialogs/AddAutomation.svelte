<script lang="ts">
    import Button,{ Icon,Label } from '@smui/button'
    import Dialog,{
    Actions,
    Content,
    Header,
    InitialFocus,
    Title
    } from '@smui/dialog'
    import IconButton from '@smui/icon-button'
    import { createEventDispatcher } from 'svelte'
    import { addAutomation,hmsLoaded,homescripts } from '../main'
    import Inputs from './Inputs.svelte'

    export let open = false

    // Event dispatcher
    const dispatch = createEventDispatcher()

    // Binded to the `Inputs.svelte` component
    let data: addAutomation = {
        days: [],
        description: '',
        enabled: true,
        homescriptId: '',
        hour: 0,
        minute: 0,
        name: '',
        timingMode: 'normal',
    }

    function reset() {
        data = {
            days: [],
            description: '',
            enabled: true,
            // `$homescripts` can be used because it is likely
            // that the user can ony invoke reset when homescripts are loaded
            homescriptId: $homescripts[0].data.id,
            hour: 0,
            minute: 0,
            name: '',
            timingMode: 'normal',
        }
        open = false
    }
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content" 
  fullscreen={$hmsLoaded && $homescripts.length > 0}>
    <Header>
        <Title id="title">
            {#if $hmsLoaded && $homescripts.length == 0}
                There are currently no Homescripts.
            {:else}
                Add Automation
            {/if}
        </Title>
        {#if $hmsLoaded && $homescripts.length > 0}
        <IconButton action="close" class="material-icons">close</IconButton>
        {/if}
        <!-- TODO: better code -->
    </Header>
    <Content id="content">
        {#if $hmsLoaded && $homescripts.length == 0}
            <p>
                You must create a Homescript in order to continue. <br /> If there are Homescripts, check that they are enabled for automations / scheduler.
                <!-- TODO: write CLI documentation and link it here -->
                <span class="text-hint"
                    >You can also use the CLI to create Homescripts. <a
                        href="https://github.com/smarthome-go/cli" target="_blank">learn more</a
                    ></span
                >
            </p>
            <Button on:click={() => (window.location.href = '/homescript')} variant="outlined">
                <Icon class="material-icons">code</Icon>
                Create one
            </Button>
        {:else}
            <Inputs bind:data />
        {/if}
    </Content>
    <Actions>
        {#if $hmsLoaded && $homescripts.length > 0}
        <Button on:click={reset}>
            <Label>Cancel</Label>
        </Button>
            <Button
                disabled={data.name == '' || data.days.length == 0}
                use={[InitialFocus]}
                on:click={() => {
                    dispatch('add', data)
                    // Reset values after creation
                    reset()
                }}
            >
                <Label>Create</Label>
            </Button>
        {:else}
        <Button>
            <Label>Dismiss</Label>
        </Button>
        {/if}
    </Actions>
</Dialog>

<style lang="scss">
    .text-hint {
        font-size: 0.9rem;
        display: block;
    }
    a {
        color: var(--clr-primary);
        opacity: 90%;
    }
</style>
