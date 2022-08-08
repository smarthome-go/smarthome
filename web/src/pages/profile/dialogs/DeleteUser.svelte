<script>
    import Button, { Label } from "@smui/button";
    import Dialog, {
        Actions,
        Content,
        InitialFocus,
        Title,
    } from "@smui/dialog";
    import Checkbox from "@smui/checkbox";
    import FormField from "@smui/form-field";
    import { createEventDispatcher } from "svelte";

    export let open;
    $: if (!open) proceed = false;

    let proceed = false;

    // Event dispatcher
    const dispatch = createEventDispatcher();
</script>

<Dialog
    bind:open
    aria-labelledby="default-focus-title"
    aria-describedby="default-focus-content"
>
    <Title id="default-focus-title">Confirm Account Deletion</Title>
    <Content id="default-focus-content">
        <span style="color: var(--clr-error)">
            This will delete all your data and your Smarthome account!
        </span>
        <br />
        <br />
        Once your account is deleted, you are logged out of Smarthome and have no
        way of logging-in again.
        <br />
        You may also want to ask your administrator to delete all backups and configuration
        exports which contain your old data.
        <br />
        <FormField>
            <Checkbox bind:checked={proceed} />
            <span slot="label"
                >I understand the consequences and want to proceed.</span
            >
        </FormField>
        <br />
        <span class="text-disabled" style="font-size: 0.9rem;"
            >Thank you for using Smarthome</span
        >
    </Content>
    <Actions>
        <Button on:click={() => dispatch("delete", null)} disabled={!proceed}>
            <Label>Confirm</Label>
        </Button>
        <Button defaultAction use={[InitialFocus]}>
            <Label>Cancel</Label>
        </Button>
    </Actions>
</Dialog>
