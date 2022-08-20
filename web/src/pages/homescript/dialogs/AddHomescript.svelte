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
    import Autocomplete from "@smui-extra/autocomplete";
    import { Text } from "@smui/list";
    import { onMount } from "svelte";
    export let open = false;

    // Input data
    let id = "";
    let name = "";
    let description = "";
    let workspace = "default";

    let newWorkspace = "";
    let workspaceText = "";
    let newWorkspaceOpen = false;

    let workspaces: string[] = [];
    $: workspaces = [
        ...new Set([
            ...$homescripts.map((h) => h.data.data.workspace),
            "default",
        ]),
    ];

    // Event dispatcher
    const dispatch = createEventDispatcher();

    function submit() {
        dispatch("add", { id, name, description, workspace });
        // Reset data after creation
        id = "";
        name = "";
        description = "";
        workspace = "default";
        open = false;
    }

    onMount(() => {
        workspace = "default";
    });
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <Header>
        <Title id="title">Add Homescript</Title>
    </Header>
    <Content id="content">
        <div class="text">
            <Autocomplete
                style="width: 100%; margin-bottom: .5rem;"
                label="Select Workspace"
                options={workspaces}
                bind:value={workspace}
                noMatchesActionDisabled={false}
                bind:text={workspaceText}
                on:SMUIAutocomplete:noMatchesAction={() => {
                    newWorkspace = workspaceText;
                    newWorkspaceOpen = true;
                }}
            >
                <div slot="no-matches">
                    <Text>Add Workspace</Text>
                </div>
            </Autocomplete>
            <br />
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
                workspace = "default";
                open = false;
            }}
        >
            <Label>Cancel</Label>
        </Button>
        <Button
            disabled={name === "" || id === ""}
            use={[InitialFocus]}
            on:click={() => {
                submit();
            }}
        >
            <Label>Create</Label>
        </Button>
    </Actions>

    <Dialog
        slot="over"
        bind:open={newWorkspaceOpen}
        aria-labelledby="workspace-dialog-title"
        aria-describedby="workspace-dialog-content"
    >
        <Title id="workspace-dialog-title">New Item</Title>
        <Content id="workspace-dialog-content">
            <Textfield bind:value={newWorkspace} label="New Workspace" />
        </Content>
        <Actions>
            <Button>
                <Label>Cancel</Label>
            </Button>
            <Button
                on:click={() => {
                    workspaces = [...workspaces, newWorkspace];
                    workspace = newWorkspace;
                }}
            >
                <Label>Add</Label>
            </Button>
        </Actions>
    </Dialog>
</Dialog>
