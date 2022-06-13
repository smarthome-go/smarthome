<script lang="ts">
    import Dialog, { Content, Header, InitialFocus, Title } from "@smui/dialog";
    import { createEventDispatcher, onMount } from "svelte";
    import Button, { Label } from "@smui/button";
    import Textfield from "@smui/textfield";
    import Switch from "@smui/switch";
    import Slider from "@smui/slider";
    import Select, { Option } from "@smui/select";
    import type { homescriptArgData, homescriptArgSubmit } from "../homescript";
    import { createSnackbar } from "../global";

    export let open: boolean = false;

    // Holds the argument list which is used to display the prompts
    export let args: homescriptArgData[];
    // Saves the index of the argument which is currently shown
    let currentArgumentIndex: number = 0;

    let currentArg: homescriptArgData = {
        argKey: "",
        homescriptId: "",
        prompt: "",
        mdIcon: "",
        inputType: "string",
        display: "type_default",
    };

    $: if (currentArgumentIndex + 1 <= args.length)
        currentArg = args[currentArgumentIndex];

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

    // Utility variables used for binding to non-string variables
    // Will be converted to the `real` string representation using change listeners
    let numberPlaceholder: number = 0;
    let booleanPlaceholder: boolean = false;

    function updateFromNumber() {
        argumentsWithValues[currentArgumentIndex].value =
            numberPlaceholder.toString();
    }
    function updateFromBoolean() {
        argumentsWithValues[currentArgumentIndex].value =
            booleanPlaceholder.toString();
    }

    $: if (currentArg.inputType == "number" && numberPlaceholder)
        updateFromNumber();

    $: if (
        currentArg.inputType == "boolean" &&
        // Used in order to trick svelte into running this every time the booleanPlaceholder changes
        (booleanPlaceholder == false || booleanPlaceholder == true)
    )
        updateFromBoolean();

    /* Switches */
    // Used for when the the `display` is set to `string_switches`
    interface SwitchResponse {
        id: string;
        name: string;
        powerOn: boolean;
        watts: number;
    }

    let switchesLoaded: boolean = false;
    let switches: SwitchResponse[] = [];

    async function loadSwitches() {
        try {
            const res = await (await fetch("/api/switch/list/personal")).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            switches = res;
            switchesLoaded = true;
        } catch (err) {
            $createSnackbar(`Could not load switches: ${err}`);
        }
    }

    onMount(() => {
        for (let arg of arguments)
            argumentsWithValues.push({ key: arg.argKey, value: "" });
    });
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <Header>
        <Title id="title">{currentArg.prompt}</Title>
    </Header>
    <Content id="content">
        <div
            class="inputs"
            class:fill={currentArg.display === "number_hour" ||
                currentArg.display === "number_minute"}
        >
            {#if currentArg.inputType === "string"}
                {#if currentArg.display === "string_switches"}
                    <Select
                        bind:value={argumentsWithValues[currentArgumentIndex]
                            .value}
                        label="Select Menu"
                    >
                        {#each fruits as fruit}
                            <Option value={fruit}>{fruit}</Option>
                        {/each}
                    </Select>
                {/if}
            {:else if currentArg.inputType === "number"}
                {#if currentArg.display === "type_default"}
                    <Textfield
                        style="width: 100%;"
                        bind:value={numberPlaceholder}
                        label={currentArg.argKey}
                        type="number"
                    />
                {:else if currentArg.display === "number_hour"}
                    <Slider
                        bind:value={numberPlaceholder}
                        min={0}
                        max={24}
                        step={1}
                        discrete
                    />
                {:else if currentArg.display === "number_minute"}
                    <Slider
                        bind:value={numberPlaceholder}
                        min={0}
                        max={60}
                        step={1}
                        discrete
                    />
                {/if}
            {:else}
                boolean
            {/if}
        </div>
        <div class="actions">
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
        </div>
    </Content>
</Dialog>

<style lang="scss">
    .inputs {
        padding: 2rem;

        // Some inputs, for example the slider require an elevated background
        &.fill {
            background-color: var(--clr-height-0-1);
            border-radius: 0.3rem;
            padding: 2rem 0;
        }
    }
    .actions {
        display: flex;
        justify-content: flex-end;
    }
</style>
