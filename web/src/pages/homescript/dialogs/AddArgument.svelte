<script lang="ts">
    import Dialog, {
        Actions,
        Content,
        Header,
        InitialFocus,
        Title,
    } from "@smui/dialog";
    import { createEventDispatcher } from "svelte";
    import Button, { Label } from "@smui/button";
    import Textfield from "@smui/textfield";
    import CharacterCounter from "@smui/textfield/character-counter";

    export let open: boolean = false;

    // Input data
    let argKey: string = "";
    let prompt: string = "";
    let inputType: "string" | "number" | "boolean" = "string";
    let display:
        | "type_default"
        | "string_switches"
        | "boolean_yes_no"
        | "boolean_on_off"
        | "number_hour"
        | "number_minute" = "type_default";

    // Event dispatcher
    const dispatch = createEventDispatcher();

    // Is called when the create button is pressed
    function submit() {
        dispatch("add", { argKey, prompt, inputType, display });
        // Reset data after creation
        argKey = "";
        prompt = "";
        inputType = "string";
        display = "type_default";
        open = false;
    }
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <Header>
        <Title id="title">Add Argument Prompt</Title>
    </Header>
    <Content id="content">
        <span class="text-hint">Identifier and prompt of the argument</span>
        <Textfield
            bind:value={argKey}
            input$maxlength={100}
            label="Key"
            required
            style="width: 100%;"
            helperLine$style="width: 100%;"
        >
            <svelte:fragment slot="helper">
                <CharacterCounter>0 / 100</CharacterCounter>
            </svelte:fragment>
        </Textfield>
        <Textfield
            bind:value={prompt}
            label="Prompt"
            required
            style="width: 100%;"
            helperLine$style="width: 100%;"
        />
    </Content>
    <Actions>
        <Button
            on:click={() => {
                argKey = "";
                prompt = "";
                inputType = "string";
                display = "type_default";
                open = false;
            }}
        >
            <Label>Cancel</Label>
        </Button>
        <Button
            disabled={argKey === "" || prompt === ""}
            use={[InitialFocus]}
            on:click={() => {
                submit();
            }}
        >
            <Label>Create</Label>
        </Button>
    </Actions>
</Dialog>
