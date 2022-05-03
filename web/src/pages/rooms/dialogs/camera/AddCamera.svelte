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
    import type { Camera } from "../../main";

    let open = false;
    export let cameras: Camera[] = [];

    let id = "";
    let name = "";
    let url = "";
    let roomId = "";

    let idDirty = false;
    let nameDirty = false;
    let urlDirty = false;

    export function show() {
        open = true;
        id = "";
        name = "";
        url = "";
        idDirty = false;
        nameDirty = false;
        urlDirty = false;
    }

    export let onAdd = (
        _id: string,
        _name: string,
        _url: string,
        _roomId: string
    ) => {};

    let idInvalid = false;
    $: idInvalid =
        (idDirty && id === "") ||
        id.includes(" ") ||
        cameras.find((s) => s.id === id) !== undefined;
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <Title id="title">Add Camera</Title>
    <Content id="content">
        <Textfield
            bind:value={id}
            bind:dirty={idDirty}
            bind:invalid={idInvalid}
            input$maxlength={50}
            label="Camera Id"
            required
        >
            <svelte:fragment slot="helper">
                <CharacterCounter>0 / 50</CharacterCounter>
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
                <CharacterCounter>0 / 50</CharacterCounter>
            </svelte:fragment>
        </Textfield>
        <Textfield bind:value={url} label="Url" />
    </Content>
    <Actions>
        <Button>
            <Label>Cancel</Label>
        </Button>
        <Button
            disabled={idInvalid || id === "" || name === "" || url === ""}
            use={[InitialFocus]}
            on:click={() => {
                onAdd(id, name, url, roomId);
            }}
        >
            <Label>Create</Label>
        </Button>
    </Actions>
</Dialog>
