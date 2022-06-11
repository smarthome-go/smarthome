<script lang="ts">
    import Switch from "@smui/switch/src/Switch.svelte";
    import Textfield from "@smui/textfield";
    import CharacterCounter from "@smui/textfield/character-counter";
    import type { homescriptArgData, homescriptData } from "../../homescript";
    import IconPicker from "../../components/IconPicker/IconPicker.svelte";
    import Button, { Label } from "@smui/button";
    import { homescripts, loading } from "./main";
    import IconButton from "@smui/icon-button";
    import AddArgument from "./dialogs/AddArgument.svelte";
    import { createSnackbar } from "../../global";
    import Argument from "./Argument.svelte";

    let iconPickerOpen = false;
    let addArgOpen: boolean = false;
    export let deleteOpen = false;

    // Can be bound in order to allow data modification
    export let data: homescriptData;

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

<div class="container">
    <!-- Names and Text -->
    <div class="text">
        <span class="text-hint">Name and Description</span>
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
    </div>
    <div class="toggles">
        <span class="text-hint">Attributes and visibility</span>
        <div class="toggles__item">
            <Switch bind:checked={data.schedulerEnabled} />
            <span class="text-hint">
                Automation selection {data.schedulerEnabled
                    ? "shown"
                    : "hidden"}
            </span>
        </div>
        <div class="right__toggles__item">
            <Switch bind:checked={data.quickActionsEnabled} />
            <span class="text-hint">
                Quick actions selection {data.quickActionsEnabled
                    ? "shown"
                    : "hidden"}
            </span>
        </div>
    </div>
    <div class="arguments">
        <span class="text-hint">Arguments</span>
        <div
            class="arguments__list"
            class:empty={$homescripts.find((h) => h.data.data.id === data.id)
                .arguments.length === 0}
        >
            {#if $homescripts.find((h) => h.data.data.id === data.id).arguments.length === 0}
                <span class="text-disabled">No arguments set up.</span>
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
    <div class="actions">
        <Button
            on:click={() => {
                iconPickerOpen = true;
            }}
        >
            Change Icon
        </Button>
        <Button on:click={() => (deleteOpen = true)}>
            <Label>Delete</Label>
        </Button>
    </div>
</div>

<style lang="scss">
    @use "../../mixins" as *;
    .container {
        display: flex;
        flex-wrap: wrap;
        gap: 2rem;

        @include not-widescreen {
            flex-direction: column;
        }
    }

    .toggles {
        background-color: var(--clr-height-1-2);
        padding: 1rem;
        border-radius: 0.3rem;

        @include widescreen {
            width: 100%;
        }

        &__item {
            @include mobile {
                span {
                    display: block;
                }
            }
        }
    }

    .arguments {
        border-radius: 0.3rem;
        padding: 0.9rem 1rem;
        background-color: var(--clr-height-1-2);
        display: block;
        width: 100%;

        &__list {
            display: flex;
            align-items: center;
            gap: 0.5rem;

            &.empty {
                margin-top: 0.5rem;
                align-items: center;
                justify-content: space-between;
            }
        }
    }

    .actions {
        width: 100%;
        display: block;
    }

    .text {
        width: 100%;
    }
</style>
