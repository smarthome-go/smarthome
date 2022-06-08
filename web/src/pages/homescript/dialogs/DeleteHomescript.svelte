<script lang="ts">
    import Button, { Label } from "@smui/button";
    import Dialog, {
        Actions,
        Content,
        InitialFocus,
        Title,
    } from "@smui/dialog";
    import { createEventDispatcher } from "svelte";
    import type { homescriptData } from "../../../homescript";

    export let open = false;
    export let data: homescriptData;

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
    <Title id="confirmation-title">Confirm Deletion of '{data.name}'</Title>
    <Content id="confirmation-content">
        Deletion can cause unwanted side effects, make shure you can identify them.
        Automations depending on this Homescript will also be removed.
        Are you shure you want to proceed?</Content
    >
    <Actions>
        <Button
            on:click={() => {
                dispatch("delete", {id: data.id});
            }}
        >
            <Label>Delete</Label>
        </Button>
        <Button use={[InitialFocus]}>
            <Label>Cancel</Label>
        </Button>
    </Actions>
</Dialog>
