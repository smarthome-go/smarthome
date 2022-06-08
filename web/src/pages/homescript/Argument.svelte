<script lang="ts">
    import type { homescriptArg } from "../../homescript";
    import EditArgument from "./dialogs/arguments/EditArgument.svelte";
    import Ripple from "@smui/ripple";
    import { loading } from "./main";
    import { createSnackbar } from "../../global";
    import { createEventDispatcher } from "svelte";

    const dispatch = createEventDispatcher();

    // Keeps track of wether the edit dialog is open
    let editOpen: boolean = false;

    // Is bound externally
    export let data: homescriptArg;

    // Sends a modification request
    async function modifyHomescriptArgument() {
        $loading = true;

        try {
            const res = await (
                await fetch("/api/homescript/arg/modify", {
                    method: "PUT",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        id: data.id,
                        argKey: data.data.argKey,
                        prompt: data.data.prompt,
                        mdIcon: data.data.mdIcon,
                        inputType: data.data.inputType,
                        display: data.data.display,
                    }),
                })
            ).json();
            if (!res.success) throw Error(res.error);
        } catch (err) {
            $createSnackbar(`Failed to modify argument: Error: ${err}`);
        }
        $loading = false;
    }
</script>

<EditArgument
    on:delete={() => {
        dispatch("delete", null);
        editOpen = false;
    }}
    on:modify={modifyHomescriptArgument}
    bind:data={data.data}
    bind:open={editOpen}
/>

<div
    class="argument"
    on:click={() => (editOpen = true)}
    use:Ripple={{ surface: true, color: "primary" }}
>
    {data.data.argKey}
    <i class="material-icons">{data.data.mdIcon}</i>
</div>

<style lang="scss">
    .argument {
        border-radius: 0.6rem;
        background-color: var(--clr-height-3-4);
        color: var(--clr-primary);
        padding: 0.3rem 0.6rem;
        font-size: 0.8rem;
        opacity: 70%;
        cursor: pointer;
        user-select: none;

        display: flex;
        align-items: center;
        gap: 0.4rem;

        i {
            font-size: 1rem;
        }
    }
</style>
