<script lang="ts">
    import Ripple from '@smui/ripple'
    import { onMount } from 'svelte'
    export let value: Date
    export let label: string

    let inputElement: HTMLInputElement

    // Approximates if the date picker is currently in use
    let active = false

    document.addEventListener(
        'click',
        (event) => {
            console.log(document.activeElement)
            active = document.activeElement === inputElement
        },
        true
    )
    onMount(() => {
        inputElement.onfocus = () => {
            active = true
        }
    })
</script>

<span>
    <input
        class="text-hint"
        use:Ripple={{ surface: true }}
        bind:value
        bind:this={inputElement}
        type="date"
        name="date"
        id="date"
    />
    <span class:active id="hint" class="text-hint">
        {label}
    </span>
</span>

<style lang="scss">
    input {
        border: none;
        outline: none;
        background-color: transparent;
        -webkit-font-smoothing: antialiased;
        font-family: Roboto, sans-serif;
        font-weight: thin;
        border-radius: 0.2rem;
        padding: 0.4rem 0.6rem;
        font-size: 1.2rem;
        border: 0.5px solid
            var(--mdc-segmented-button-outline-color, rgba(255, 255, 255, 0.12));
        cursor: pointer;
    }

    input:focus {
        color: var(--text-high);
    }

    input::-webkit-calendar-picker-indicator {
        color: transparent;
        background: none;
        z-index: 1;
        cursor: pointer;
    }

    input:before {
        background: none;
        font-family: 'Material Icons';
        content: 'event';
        font-size: 0.9rem;
        width: 1rem;
        height: 1rem;
        position: relative;
        left: 80%;
        opacity: 1;
        color: var(--text-hint);
        padding-top: 0.1rem;
        padding-left: 0.5rem;
        box-sizing: border-box;
        transition: 0.1s;
    }

    input:focus::before {
        color: var(--clr-primary);
    }

    #hint {
        font-size: 0.75rem;
        margin-left: 0.1rem;
        margin-top: 0.1rem;
        -webkit-font-smoothing: antialiased;
        display: block;
        opacity: 0;
        transition: opacity 150ms 0ms linear; // From other mdc components
    }

    #hint.active {
        opacity: 1;
    }
</style>