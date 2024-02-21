<script lang='ts'>
    import Button, { Label } from "@smui/button";
    import Dialog, {
        Actions,
        Content,
        InitialFocus,
        Title,
    } from "@smui/dialog";
    import Textfield from "@smui/textfield";
    import CharacterCounter from "@smui/textfield/character-counter";
    import HelperText from "@smui/textfield/helper-text";
    import { createEventDispatcher } from "svelte";
    import type { CreateDriver } from "../driver";

    export let open: boolean;

    let input: CreateDriver = {
       name: "",
       vendorId: "",
       modelId: "",
       version: "",
       homescriptCode: ""
    }

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
            bind:value={input.name}
            label="Name"
            style="width: 100%;"
        >
            <svelte:fragment slot="helper">
                <HelperText>A short and fitting name</HelperText>
            </svelte:fragment>
        </Textfield>
        <Textfield
            bind:value={input.vendorId}
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
            bind:value={input.modelId}
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
            bind:value={input.version}
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
                dispatch("create", structuredClone(input))
            input.name = ""
            input.vendorId = ""
            input.modelId = ""
            input.version = ""
            }}
            disabled={
                isEmpty(input.name) ||
                isEmpty(input.vendorId) ||
                isEmpty(input.modelId) ||
                isEmpty(input.version)
            }
        >
            <Label>Create</Label>
        </Button>
    </Actions>
</Dialog>
