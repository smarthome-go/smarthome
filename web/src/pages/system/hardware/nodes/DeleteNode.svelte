<script>
    import Button, { Label } from '@smui/button'
    import Dialog, { Actions, Content, InitialFocus, Title } from '@smui/dialog'
    import { createEventDispatcher } from 'svelte'

    export let open = false

    // Event dispatcher
    const dispatch = createEventDispatcher()
</script>

<Dialog bind:open aria-labelledby="delete-title" aria-describedby="delete-content">
    <Title id="delete-title">Confirm Deletion</Title>
    <Content id="delete-content">
        <span class="text-hint">
            <ul class="consequences">
                <li>
                    Using <code>None</code> as the target node for affected switches.
                </li>
                <li>
                    Increased latency for some switches.
                </li>
                <li>
                    Service outage for some switches.
                </li>
            </ul>
            Please only proceed if you are able to identify all consequences of your action. Are you
            sure that you want to proceed?
        </span>
    </Content>
    <Actions>
        <Button on:click={() => dispatch('delete', null)}>
            <Label>Delete</Label>
        </Button>
        <Button defaultAction use={[InitialFocus]}>
            <Label>Cancel</Label>
        </Button>
    </Actions>
</Dialog>
