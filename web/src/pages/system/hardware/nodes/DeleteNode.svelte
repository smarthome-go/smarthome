<script>
    import Button, { Label } from "@smui/button";
    import Dialog, {
        Actions,
        Content,
        InitialFocus,
        Title,
    } from "@smui/dialog";
    import { createEventDispatcher } from "svelte";

    export let open = false;

    // Event dispatcher
    const dispatch = createEventDispatcher();
</script>

<Dialog
    bind:open
    aria-labelledby="delete-title"
    aria-describedby="delete-content"
>
    <Title id="delete-title">Confirm Deletion</Title>
    <Content id="delete-content">
        <span class="text-hint">
            You are about to delete a hardware node.
            Users issuing power requests may experience an increased amount of latency.
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
