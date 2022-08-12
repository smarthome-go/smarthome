<script lang="ts">
    import Button, { Label } from "@smui/button";
    import Dialog, {
        Actions,
        Content,
        InitialFocus,
        Title,
    } from "@smui/dialog";
    import FormField from "@smui/form-field";
    import Switch from "@smui/switch";
    import Textfield from "@smui/textfield";
    import CharacterCounter from "@smui/textfield/character-counter";
    import HelperText from "@smui/textfield/helper-text";
    import { createEventDispatcher, onMount } from "svelte";
    import type { hardwareNode } from "../types";

    // Event dispatcher
    const dispatch = createEventDispatcher();

    // Whether the dialog should be open or closed
    export let open = false;

    // Input data
    export let data: hardwareNode = {
        url: "",
        name: "",
        token: "",
        enabled: false,
        online: false,
    };

    // Is bound to the inputs: on a successful edit, the component's data is updated
    let name = "";
    let token = "";
    let enabled = true;

    onMount(() => {
        name = data.name;
        token = data.token;
        enabled = data.enabled;
    });
</script>

<Dialog
    bind:open
    aria-labelledby="create-node-title"
    aria-describedby="create-node-content"
>
    <Title id="create-node-title">Edit Hardware Node</Title>
    <Content id="create-node-content">
        <Textfield
            bind:value={name}
            label="Name"
            input$maxlength={30}
            style="width: 100%;"
        >
            <svelte:fragment slot="helper">
                <HelperText>A short and fitting name</HelperText>
                <CharacterCounter>0 / 30</CharacterCounter>
            </svelte:fragment>
        </Textfield>
        <Textfield
            bind:value={token}
            label="Token"
            type="password"
            input$maxlength={100}
            style="width: 100%;"
        >
            <svelte:fragment slot="helper">
                <HelperText>The node's authentication token</HelperText>
                <CharacterCounter>0 / 100</CharacterCounter>
            </svelte:fragment>
        </Textfield>
        <FormField>
            <Switch bind:checked={enabled} />
            <span slot="label">Device {enabled ? "Enabled" : "Disabled"}</span>
        </FormField>
    </Content>
    <Actions>
        <Button
            on:click={() => {
                name = data.name;
                token = data.token;
                enabled = data.enabled;
            }}
        >
            <Label>Cancel</Label>
        </Button>
        <Button
            defaultAction
            use={[InitialFocus]}
            on:click={() =>
                dispatch("edit", {
                    name,
                    token,
                    enabled,
                })}
            disabled={name === "" ||
                (name === data.name &&
                    token === data.token &&
                    enabled === data.enabled)}
        >
            <Label>Edit</Label>
        </Button>
    </Actions>
</Dialog>
