<script lang="ts">
    import Button, { Icon, Label } from '@smui/button'
    import Dialog, { Actions, Content, Header, InitialFocus, Title } from '@smui/dialog'
    import IconButton from '@smui/icon-button'
    import { createEventDispatcher } from 'svelte'
    import { hmsLoaded, homescripts, sunTimes } from '../main'
    import type { addAutomation } from '../main'
    import Inputs from './Inputs.svelte'

    export let open = false

    // Event dispatcher
    const dispatch = createEventDispatcher()

    // Is required in order to reset the previous day-state
    let selectedDays: string[] = []

    // Is bound to the `Inputs.svelte` component
    let data: addAutomation = {
        days: [],
        description: '',
        enabled: true,
        homescriptId: '',
        hour: 0,
        minute: 0,
        name: '',
        trigger: 'cron',
    }

    function reset() {
        // Reset the reverse-bound days
        selectedDays = []
        data = {
            days: [],
            description: '',
            enabled: true,
            // `$homescripts` can be used because it is likely
            // that the user can only invoke reset when Homescripts are loaded
            homescriptId: $homescripts[0].data.id,
            hour: 0,
            minute: 0,
            name: '',
            trigger: 'cron',
        }
        open = false
    }

    // Show the correct time in the time picker when using sun-times
    $: if (data.trigger == 'on_sunrise' || data.trigger == 'on_sunset') {
        switch (data.trigger) {
            case 'on_sunrise':
                data.hour = $sunTimes.sunriseHour
                data.minute = $sunTimes.sunriseMinute
                break
            case 'on_sunset':
                data.hour = $sunTimes.sunsetHour
                data.minute = $sunTimes.sunsetMinute
                break
        }
    }
</script>

<Dialog
    bind:open
    aria-labelledby="title"
    aria-describedby="content"
    fullscreen={$hmsLoaded && $homescripts.length > 0}
>
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
        <!-- TODO: fix this ugly code -->
    </Header>
    <Content id="content">
        {#if $hmsLoaded && $homescripts.length == 0}
            <p>
                You must create a Homescript in order to continue. <br /> If there are Homescripts, check
                that they are enabled for automations / scheduler.
            </p>
            <Button on:click={() => (window.location.href = '/homescript')} variant="outlined">
                <Icon class="material-icons">code</Icon>
                Create one
            </Button>
        {:else}
            <Inputs bind:data bind:selectedDays />
        {/if}
    </Content>
    <Actions>
        {#if $hmsLoaded && $homescripts.length > 0}
            <Button on:click={reset}>
                <Label>Cancel</Label>
            </Button>
            <Button
                disabled={data.name === '' ||
                    (data.trigger === 'cron' && data.days.length === 0) ||
                    (data.trigger === 'interval' &&
                        (data.triggerInterval <= 0 || data.triggerInterval > 60 * 60 * 24 * 365))}
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
