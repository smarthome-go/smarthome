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
    import { homescripts } from "../main";
    export let open = false;

    // Input data
    let id = "";
    let name = "";
    let description = "";

    // Event dispatcher
    const dispatch = createEventDispatcher();

    function submit() {
        dispatch("add", { id: id, name: name, description: description });
        // Reset data after creation
        id = "";
        name = "";
        description = "";
        open = false;
    }
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <Header>
        <Title id="title">Add Homescript</Title>
    </Header>
    <Content id="content">
        <div class="text">
            <span class="text-hint">Name and description of the Homescript</span
            >
            <Textfield
                bind:value={id}
                invalid={id.includes(" ") ||
                    $homescripts.find((h) => h.data.data.id === id) !==
                        undefined}
                input$maxlength={30}
                label="Id"
                required
                style="width: 100%;"
                helperLine$style="width: 100%;"
            >
                <svelte:fragment slot="helper">
                    <CharacterCounter>0 / 30</CharacterCounter>
                </svelte:fragment>
            </Textfield>
            <Textfield
                bind:value={name}
                input$maxlength={30}
                label="Name"
                required
                style="width: 100%;"
                helperLine$style="width: 100%;"
            >
                <svelte:fragment slot="helper">
                    <CharacterCounter>0 / 30</CharacterCounter>
                </svelte:fragment>
            </Textfield>
            <Textfield
                bind:value={description}
                label="Description"
                style="width: 100%;"
                helperLine$style="width: 100%;"
            />
        </div>
    </Content>
    <Actions>
        <Button
            on:click={() => {
                id = "";
                name = "";
                description = "";
                open = false;
            }}
        >
            <Label>Cancel</Label>
        </Button>
        <Button
            disabled={name === ""}
            use={[InitialFocus]}
            on:click={() => {
                submit();
            }}
        >
            <Label>Create</Label>
        </Button>
    </Actions>
</Dialog>
