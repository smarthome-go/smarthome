<script lang="ts">
    import Checkbox from "@smui/checkbox";

    import type { reminder } from "./types";
    import { createSnackbar } from "../../../global";
    import { createEventDispatcher } from "svelte";

    const dispatch = createEventDispatcher();

    export let data: reminder;

    let checked = false;
    $: if (checked) deleteSelf();

    async function deleteSelf() {
        try {
            const res = await (
                await fetch("/api/reminder/delete", {
                    headers: { "Content-Type": "application/json" },
                    method: "DELETE",
                    body: JSON.stringify({ id: data.id }),
                })
            ).json();
            if (!res.success) throw Error();
            setTimeout(() => {
                dispatch("delete", null);
            }, 10);
        } catch (err) {
            $createSnackbar("Could not mark reminder as completed");
        }
    }

    const priorities = [
        { label: "LOW", color: "var(--clr-priority-low)" },
        { label: "NORMAL", color: "var(--clr-success)" },
        { label: "MEDIUM", color: "var(--clr-priority-medium)" },
        { label: "HIGH", color: "var(--clr-warn)" },
        { label: "URGENT", color: "var(--clr-error)" },
    ];
</script>

<div
    class="reminder mdc-elevation--z3"
    style:--clr-border={priorities[data.priority].color}
>
    <div class="reminder__left">
        <span class="reminder__left__title">{data.name}</span>
        <span class="reminder__left__description">{data.description}</span>
    </div>
    <Checkbox bind:checked />
</div>

<style lang="scss">
    .reminder {
        background-color: var(--clr-height-2-3);
        border-radius: 0.2rem;
        padding: 0.5rem 1rem;
        display: flex;
        align-items: center;
        justify-content: space-between;
        border-left: solid var(--clr-border) 0.3rem;

        &__left {
            width: calc(100% - 4rem);

            &__title {
                font-weight: bold;
                font-size: .9rem;
            }

            &__description {
                display: block;
                font-size: 0.75rem;

                white-space: nowrap;
                text-overflow: ellipsis;
                overflow: hidden;
            }
        }
    }
</style>
