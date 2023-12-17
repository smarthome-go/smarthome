<script lang='ts'>
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
    export let open: boolean;

    // Input data
    let name = "";
    let vendorId = "";
    let modelId = "";
    let version = "";

    // Event dispatcher
    const dispatch = createEventDispatcher();

    function isEmpty(input: string): boolean {
        return input === ""  || input.replaceAll(" ", "") === ""
    }
</script>

<Dialog
    bind:open
    aria-labelledby="create-node-title"
    aria-describedby="create-node-content"
>
    <Title id="create-node-title">Create Hardware Driver</Title>
    <Content id="create-node-content">
        <Textfield
            bind:value={name}
            label="Name"
            style="width: 100%;"
        >
            <svelte:fragment slot="helper">
                <HelperText>A short and fitting name</HelperText>
            </svelte:fragment>
        </Textfield>
        <Textfield
            bind:value={vendorId}
            label="Vendor ID"
            input$maxlength={50}
            style="width: 100%;"
        >
            <svelte:fragment slot="helper">
                <HelperText>Vendor ID</HelperText>
                <CharacterCounter>0 / 50</CharacterCounter>
            </svelte:fragment>
        </Textfield>
        <Textfield
            bind:value={modelId}
            label="Model ID"
            input$maxlength={50}
            style="width: 100%;"
        >
            <svelte:fragment slot="helper">
                <HelperText>Vendor ID</HelperText>
                <CharacterCounter>0 / 50</CharacterCounter>
            </svelte:fragment>
        </Textfield>
        <Textfield
            bind:value={version}
            label="Version"
            input$maxlength={50}
            style="width: 100%;"
        >
            <svelte:fragment slot="helper">
                <HelperText>Version</HelperText>
                <CharacterCounter>0 / 50</CharacterCounter>
            </svelte:fragment>
        </Textfield>
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
                    name,
                    vendorId,
                    modelId,
                    version,
            })
            name = ""
            vendorId = ""
            modelId = ""
            version = ""
            }}
            disabled={isEmpty(name) || isEmpty(vendorId) || isEmpty(modelId) || isEmpty(version) }
        >
            <Label>Create</Label>
        </Button>
    </Actions>
</Dialog>
