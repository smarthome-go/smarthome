<script lang="ts">
    import Ripple from "@smui/ripple";
    import { createEventDispatcher } from "svelte";
    import { v4 as uuidv4 } from "uuid";
    export let value = "#ffffff";

    const id = `${new Date().getTime()}-${uuidv4()}`;
    const dispatch = createEventDispatcher<{ change: { value: string } }>();

    function handleChange(event: Event) {
        const target = event.target as HTMLInputElement;
        dispatch("change", { value: target.value });
    }
</script>

<span>
    <input bind:value type="color" name="color" {id} on:change={handleChange} />
    <label use:Ripple={{ surface: true }} for={id}>PICK COLOR</label>
</span>

<style lang="scss">
    input {
        opacity: 0;
        width: 0;
        height: 0;
    }
    label {
        -webkit-font-smoothing: antialiased;
        font-family: Roboto, sans-serif;
        font-weight: 500;
        border-radius: 0.2rem;
        padding: 0.36rem 0.7rem;
        font-size: 14px;
        border: 0.5px solid
            var(--mdc-segmented-button-outline-color, rgba(255, 255, 255, 0.12));
        cursor: pointer;

        display: block;
        color: var(--clr-primary);
    }
    span {
        display: flex;
        align-items: center;
    }
</style>
