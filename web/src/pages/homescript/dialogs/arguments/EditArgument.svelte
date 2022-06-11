<script lang="ts">
    import Button, { Label } from "@smui/button";
    import Dialog, {
        Actions,
        Content,
        Header,
        InitialFocus,
        Title,
    } from "@smui/dialog";
    import IconButton from "@smui/icon-button";
    import type { homescriptArgData } from "src/homescript";
    import { createEventDispatcher } from "svelte";
    import Textfield from "@smui/textfield";
    import CharacterCounter from "@smui/textfield/character-counter";
    import { onMount } from "svelte";
    import IconPicker from "../../../../components/IconPicker/IconPicker.svelte";
    import ArgumentInputs from "../../ArgumentInputs.svelte";

    export let open: boolean = false;

    // This is required as an override for the additional icon popup
    // TODO: reevaluate this choice
    $: document.body.style.overflow = open ? "hidden" : "auto";

    let pickIconOpen: boolean = false;

    // Event dispatcher
    const dispatch = createEventDispatcher();

    // Only bound externally in order to use preset values
    export let data: homescriptArgData;
    let dataChanged: boolean = false;

    // Internal values which keep track of change
    let argKeyBefore: string;
    let promptBefore: string;
    let mdIconBefore: string;
    let inputTypeBefore: "string" | "number" | "boolean";
    let displayBefore:
        | "type_default"
        | "string_switches"
        | "boolean_yes_no"
        | "boolean_on_off"
        | "number_hour"
        | "number_minute";

    // Updates whether `dataChanged` or not
    $: if (data) updateDataChanged();
    function updateDataChanged() {
        dataChanged =
            argKeyBefore !== data.argKey ||
            promptBefore !== data.prompt ||
            mdIconBefore !== data.mdIcon ||
            inputTypeBefore !== data.inputType ||
            displayBefore !== data.display;
    }

    // Restores any changes made to the data based on the previous saves
    // Is used when the cancel button is pressed
    function resetChanges() {
        data.argKey = argKeyBefore;
        data.prompt = promptBefore;
        data.mdIcon = mdIconBefore;
        data.inputType = inputTypeBefore;
        data.display = displayBefore;
    }

    // Updates the previously saved changes to match the current ones
    // Is used when the submit / edit button is pressed
    function updateBeforeData() {
        argKeyBefore = data.argKey;
        promptBefore = data.prompt;
        mdIconBefore = data.mdIcon;
        inputTypeBefore = data.inputType;
        displayBefore = data.display;
    }

    onMount(updateBeforeData);
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <IconPicker
        bind:open={pickIconOpen}
        bind:selected={data.mdIcon}
        slot="over"
    />
    <Header>
        <Title id="title">Edit Argument</Title>
    </Header>
    <Content id="content">
        <ArgumentInputs bind:data />
        <div class="actions">
            <div>
                <Button
                    on:click={() => {
                        dispatch("delete", null);
                    }}
                >
                    <Label>Delete</Label>
                </Button>
            </div>
            <div>
                <Button
                    variant="outlined"
                    on:click={() => {
                        pickIconOpen = true;
                    }}
                >
                    <Label>Change icon</Label>
                </Button>
            </div>
        </div>
    </Content>
    <Actions>
        <Button on:click={resetChanges}>
            <Label>Cancel</Label>
        </Button>
        <Button
            disabled={!dataChanged}
            use={[InitialFocus]}
            on:click={() => {
                dispatch("modify", { data });
            }}
        >
            <Label>Edit</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
    @use "../../../../mixins" as *;
    .actions {
        display: flex;
        gap: 2rem;
        align-items: center;
        border-radius: 0.3rem;
        margin: 0.5rem 0;

        @include mobile {
            gap: 1rem;
            flex-direction: row-reverse;

            div {
                width: auto;
            }
        }
    }
</style>
