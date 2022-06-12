<script lang="ts">
    import Dialog, {
        Actions,
        Content,
        Header,
        InitialFocus,
        Title,
    } from "@smui/dialog";
    import { createEventDispatcher, onMount } from "svelte";
    import Button, { Label } from "@smui/button";
    import Textfield from "@smui/textfield";
    import type { homescriptArgData, homescriptArgSubmit } from "../homescript";

    export let open: boolean = false;

    // Holds the argument list which is used to display the prompts
    export let args: homescriptArgData[];
    // Saves the index of the argument which is currently shown
    let currentArgumentIndex: number = 0;

    let currentPrompt: string = "";
    $: if (currentArgumentIndex + 1 <= args.length)
        currentPrompt = args[currentArgumentIndex].prompt;

    // Is produced when the final submit button is pressed
    let argumentsWithValues: homescriptArgSubmit[] = [];

    // Event dispatcher
    const dispatch = createEventDispatcher();

    // Is called when the submit button is pressed
    // If the button is pressed the last time, the event dispatcher dispatches the 'submit' event
    function submit() {
        if (currentArgumentIndex + 1 === args.length) {
            argumentsWithValues[currentArgumentIndex].value = "test";
            currentArgumentIndex = 0;
            dispatch("submit", argumentsWithValues);
            open = false;
            return;
        }
        argumentsWithValues[currentArgumentIndex].value = "test";
        currentArgumentIndex++;
    }

    onMount(() => {
        for (let arg of arguments)
            argumentsWithValues.push({ key: arg.argKey, value: "" });
    });
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <Header>
        <Title id="title">{currentPrompt}</Title>
    </Header>
    <Content id="content">
        <span>{currentArgumentIndex}</span>
        <Button
            on:click={() => {
                argumentsWithValues = [];
                currentArgumentIndex = 0;
                open = false;
            }}
        >
            <Label>Cancel</Label>
        </Button>
        <Button use={[InitialFocus]} on:click={submit}>
            <Label>Submit</Label>
        </Button>
    </Content>
</Dialog>
