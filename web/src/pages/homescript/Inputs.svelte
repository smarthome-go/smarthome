<script lang="ts">
    import Switch from "@smui/switch/src/Switch.svelte";
    import Textfield from "@smui/textfield";
    import CharacterCounter from "@smui/textfield/character-counter";
    import type { homescriptArgData, homescriptData } from "../../homescript";
    import IconPicker from "../../components/IconPicker/IconPicker.svelte";
    import Button, { Icon, Label } from "@smui/button";
    import { homescripts, loading } from "./main";
    import IconButton from "@smui/icon-button";
    import AddArgument from "./dialogs/AddArgument.svelte";
    import { createSnackbar } from "../../global";
    import Argument from "./Argument.svelte";
    import Autocomplete from "@smui-extra/autocomplete";
    import { Text } from "@smui/list";
    import Dialog, { Actions, Content, Title } from "@smui/dialog";

    let iconPickerOpen = false;
    let addArgOpen = false;
    export let deleteOpen = false;

    // Can be bound in order to allow data modification
    export let data: homescriptData;

    let workspaceInputText = "";
    let newWorkspaceInputText = "";
    let newWorkspaceOpen = false;

    let workspaces: string[] = [];
    $: workspaces = [
        ...new Set([
            ...$homescripts.map((h) => h.data.data.workspace),
            "default",
        ]),
    ];

    async function createHomescriptArg(
        key: string,
        prompt: string,
        inputType: "string" | "number" | "boolean",
        display:
            | "type_default"
            | "string_switches"
            | "boolean_yes_no"
            | "boolean_on_off"
            | "number_hour"
            | "number_minute" = "type_default"
    ) {
        $loading = true;
        let payload: homescriptArgData = {
            argKey: key,
            homescriptId: data.id,
            prompt: prompt,
            mdIcon: "data_array",
            inputType: inputType,
            display: display,
        };
        // Is required because the user may change the active script while this function is loading
        const selectionBefore = data.id;
        try {
            const res = await (
                await fetch("/api/homescript/arg/add", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify(payload),
                })
            ).json();
            if (!res.response.success) throw Error(res.response.error);
            // If successful, append the argument to the argument list of the current Homescript
            const selectionIndex = $homescripts.findIndex(
                (h) => h.data.data.id === selectionBefore
            );
            $homescripts[selectionIndex].arguments = [
                ...$homescripts[selectionIndex].arguments,
                {
                    id: res.id,
                    data: payload,
                },
            ];
        } catch (err) {
            $createSnackbar(`Could not create Homescript argument: ${err}`);
        }
        $loading = false;
    }

    // Requests deletion of a Homescript argument
    async function deleteHomescriptArgument(id: number) {
        $loading = true;
        try {
            let res = await (
                await fetch("/api/homescript/arg/delete", {
                    method: "DELETE",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ id }),
                })
            ).json();
            if (!res.success) throw Error(res.error);
            // Remove the deleted argument from the argument list
            const modifyIndex = $homescripts.findIndex(
                (h) => h.data.data.id === data.id
            );
            $homescripts[modifyIndex].arguments = $homescripts[
                modifyIndex
            ].arguments.filter((a) => a.id !== id);
        } catch (err) {
            $createSnackbar(`Could not delete Homescript argument: ${err}`);
        }
        $loading = false;
    }
</script>

<IconPicker bind:open={iconPickerOpen} bind:selected={data.mdIcon} />

<AddArgument
    on:add={(event) => {
        createHomescriptArg(
            event.detail.argKey,
            event.detail.prompt,
            event.detail.inputType,
            event.detail.display
        );
    }}
    bind:open={addArgOpen}
/>

<Dialog
    bind:open={newWorkspaceOpen}
    aria-labelledby="workspace-dialog-title"
    aria-describedby="workspace-dialog-content"
>
    <Title id="workspace-dialog-title">New Item</Title>
    <Content id="workspace-dialog-content">
        <Textfield bind:value={newWorkspaceInputText} label="New Workspace" />
    </Content>
    <Actions>
        <Button>
            <Label>Cancel</Label>
        </Button>
        <Button
            on:click={() => {
                workspaces = [...workspaces, newWorkspaceInputText];
                data.workspace = newWorkspaceInputText;
            }}
        >
            <Label>Add</Label>
        </Button>
    </Actions>
</Dialog>

<div class="container">
    <!-- Names and Text -->
    <div class="text">
        <Textfield
            bind:value={data.name}
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
            bind:value={data.description}
            label="Description"
            style="width: 100%;"
            helperLine$style="width: 100%;"
        />
        <br />
        <br />
        <Autocomplete
            style="width: 100%"
            label="Select Workspace"
            options={workspaces}
            bind:value={data.workspace}
            noMatchesActionDisabled={false}
            bind:text={workspaceInputText}
            on:SMUIAutocomplete:noMatchesAction={() => {
                newWorkspaceInputText = workspaceInputText;
                newWorkspaceOpen = true;
            }}
        >
            <div slot="no-matches">
                <Text>Add Workspace</Text>
            </div>
        </Autocomplete>
    </div>
    <div class="toggles-actions">
        <!-- Toggles -->
        <div class="toggles-actions__toggles">
            <span class="text-hint">Selection and visibility</span>
            <div>
                <Switch bind:checked={data.schedulerEnabled} />
                <span class="text-hint">Show Selection</span>
            </div>
            <div>
                <Switch bind:checked={data.quickActionsEnabled} />
                <span class="text-hint">Quick actions </span>
            </div>
        </div>
        <!-- Action buttons -->
        <div class="toggles-actions__actions">
            <span class="text-hint">Actions and theming</span>
            <Button
                on:click={() => {
                    iconPickerOpen = true;
                }}
            >
                Pick Icon
            </Button>
            <Button on:click={() => (deleteOpen = true)}>
                <Label>Delete</Label>
                <Icon class="material-icons">delete</Icon>
            </Button>
        </div>
    </div>
    {#if $homescripts.find((h) => h.data.data.id === data.id) !== undefined}
        <div class="arguments">
            <span class="text-hint">Argument Prompts</span>
            <div
                class="arguments__list"
                class:empty={$homescripts.find(
                    (h) => h.data.data.id === data.id
                ).arguments.length === 0}
            >
                {#if $homescripts.find((h) => h.data.data.id === data.id).arguments.length === 0}
                    <span class="text-disabled"
                        >No argument prompts set up.</span
                    >
                    <IconButton
                        class="material-icons"
                        on:click={() => (addArgOpen = true)}>add</IconButton
                    >
                {:else}
                    {#each $homescripts.find((h) => h.data.data.id === data.id).arguments as arg (arg.id)}
                        <Argument
                            on:delete={() => {
                                deleteHomescriptArgument(arg.id);
                            }}
                            bind:data={arg}
                        />
                    {/each}
                    <div class="argument">
                        <IconButton
                            class="material-icons"
                            on:click={() => (addArgOpen = true)}>add</IconButton
                        >
                    </div>
                {/if}
            </div>
        </div>
    {/if}
</div>

<style lang="scss">
    @use "../../mixins" as *;
    .container {
        display: flex;
        flex-wrap: wrap;
        gap: 2rem;
        flex-direction: column;
    }

    .toggles-actions {
        background-color: var(--clr-height-1-2);
        padding: 1rem;
        border-radius: 0.3rem;
        display: flex;
        justify-content: space-between;

        @include widescreen {
            width: 100%;
            box-sizing: border-box;
        }

        @include mobile {
            flex-direction: column;
            gap: 2rem;
        }

        &__toggles {
            div {
                span {
                    @include mobile {
                        font-size: 0.9rem;
                    }
                }
            }
        }

        &__actions {
            display: flex;
            flex-direction: column;
            align-items: start;

            span {
                margin-bottom: 0.4rem;
            }
        }
    }

    .arguments {
        border-radius: 0.3rem;
        padding: 0.7rem 1rem;
        padding-top: 0.9rem;
        background-color: var(--clr-height-1-2);
        display: block;
        min-height: 5rem;

        @include widescreen {
            width: 100%;
            box-sizing: border-box;
        }

        &__list {
            display: flex;
            align-items: center;
            flex-wrap: wrap;
            margin-top: 0.5rem;
            gap: 0.5rem;

            &.empty {
                align-items: center;
                justify-content: space-between;
            }
        }
    }

    .text {
        width: 100%;
    }
</style>
