<script>
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
    import { createEventDispatcher } from "svelte";

    // Whether the dialog should be open or closed
    export let open;

    // Input data
    let url = "";
    let name = "";
    let token = "";
    let enabled = true;

    // Event dispatcher
    const dispatch = createEventDispatcher();
</script>

<Dialog
    bind:open
    aria-labelledby="create-node-title"
    aria-describedby="create-node-content"
>
    <Title id="create-node-title">Create Hardware Node</Title>
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
            bind:value={url}
            label="Url"
            type='url'
            input$maxlength={50}
            style="width: 100%;"
        >
            <svelte:fragment slot="helper">
                <HelperText>The node's URL</HelperText>
                <CharacterCounter>0 / 50</CharacterCounter>
            </svelte:fragment>
        </Textfield>
        <Textfield
            bind:value={token}
            label="Token"
            type='password'
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
        <Button>
            <Label>Cancel</Label>
        </Button>
        <Button
            defaultAction
            use={[InitialFocus]}
            on:click={() => {
                dispatch("create", {
                    url,
                    name,
                    token,
                    enabled,
            })
            url = ""
            name = ""
            token = ""
            enabled = true
            }}
            disabled={name === "" || url === ""}
        >
            <Label>Create</Label>
        </Button>
    </Actions>
</Dialog>
