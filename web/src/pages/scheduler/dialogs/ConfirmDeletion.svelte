<script lang="ts">
    import Button, { Label } from "@smui/button";
    import Dialog, {
        Actions,
        Content,
        InitialFocus,
        Title,
    } from "@smui/dialog";
    import { createEventDispatcher } from "svelte";

    export let open: boolean = false;
    export let name: string;

    // Event dispatcher
    const dispatch = createEventDispatcher();
</script>

<!-- Deletion confirmation dialog -->
<Dialog
    bind:open
    aria-labelledby="confirmation-title"
    aria-describedby="confirmation-content"
    slot="over"
>
    <Title id="confirmation-title">Confirm Deletion of '{name}'</Title>
    <Content id="confirmation-content">
        This will abort this schedule's execution.
        <br />
        Are you sure you want to proceed?
    </Content>
    <Actions>
        <Button on:click={() => dispatch("confirm")}>
            <Label>Delete</Label>
        </Button>
        <Button use={[InitialFocus]}>
            <Label>Cancel</Label>
        </Button>
    </Actions>
</Dialog>
