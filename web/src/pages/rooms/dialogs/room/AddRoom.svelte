<script lang="ts">
    import Button, { Label } from "@smui/button";
    import Dialog, {
        Actions,
        Content,
        InitialFocus,
        Title,
    } from "@smui/dialog";
    import Textfield from "@smui/textfield";
    import CharacterCounter from "@smui/textfield/character-counter";
    import type { Room } from "../../main";

    let open = false;
    export let blacklist: Room[] = [];

    let id = "";
    let name = "";
    let description = "";

    let idDirty = false;
    let nameDirty = false;

    export function show() {
        open = true;
        id = "";
        name = "";
        description = "";
        idDirty = false;
        nameDirty = false;
    }

    export let onAdd: (
        _id: string,
        _name: string,
        _description: string
    ) => Promise<void>;

    let idInvalid = false;
    $: idInvalid =
        (idDirty && id === "") ||
        id.includes(" ") ||
        blacklist.find((r) => r.data.id === id) !== undefined;
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <Title id="title">Add Room</Title>
    <Content id="content">
        <Textfield
            bind:value={id}
            bind:dirty={idDirty}
            bind:invalid={idInvalid}
            input$maxlength={30}
            label="Room Id"
            required
        >
            <svelte:fragment slot="helper">
                <CharacterCounter>0 / 25</CharacterCounter>
            </svelte:fragment>
        </Textfield>
        <Textfield
            bind:value={name}
            bind:dirty={nameDirty}
            input$maxlength={50}
            label="Name"
            required
        >
            <svelte:fragment slot="helper">
                <CharacterCounter>0 / 45</CharacterCounter>
            </svelte:fragment>
        </Textfield>
        <Textfield bind:value={description} label="Description" />
    </Content>
    <Actions>
        <Button>
            <Label>Cancel</Label>
        </Button>
        <Button
            disabled={idInvalid || id === "" || name === ""}
            use={[InitialFocus]}
            on:click={() => {
                onAdd(id, name, description);
            }}
        >
            <Label>Create</Label>
        </Button>
    </Actions>
</Dialog>
