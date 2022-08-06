<script lang="ts">
    import Ripple from '@smui/ripple'
    import { onMount } from 'svelte'

    /** Usage detection: Approximates if the component is currently in use / active */
    let active = false
    let inputElement: HTMLInputElement // Needed for detecting usage

    document.addEventListener(
        // Used for displaying and hiding the label / helper text
        'click',
        () => {
            active = document.activeElement === inputElement
        },
        true
    )

    // Bindable values
    export let value = new Date()
    export let helperText: string
    // Will be displayed instead of the helper text if invalid is set to true
    export let invalidText: string
    // If set to true, a warning will be displayed
    export let invalid = false

    // Clears the input field and resets the value
    export function clear() {
        value = new Date()
        inputElement.value = ''
    }

    $: {
        if (
            inputElement !== null &&
            inputElement !== undefined &&
            value !== null &&
            value != undefined
        ) {
            // If the date picker is created with a predefined value, it is set here
            inputElement.value = `${value.getFullYear()}-${(Number(value.getMonth() + 1))
                .toString()
                .padStart(2, '0')}-${value
                .getDate()
                .toString()
                .padStart(2, '0')}`
        }
    }

    onMount(() => {
        inputElement.onfocus = () => {
            // Always show the helper text when the input is focused
            active = true
        }
        inputElement.oninput = () => {
            // Needed because binding to value is not optimal
            value =
                inputElement != undefined
                    ? inputElement.valueAsDate
                    : new Date()
        }
        inputElement.onchange = () => {
            value =
                inputElement != undefined
                    ? inputElement.valueAsDate
                    : new Date()
        }
    })
</script>

<span>
    <input
        class="text-hint"
        use:Ripple={{ surface: true }}
        bind:this={inputElement}
        class:invalid
        type="date"
        name="date"
        id="date"
    />
    <span class:invalid class:active id="hint" class="text-hint">
        {invalid ? invalidText : helperText}
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
        padding: 0.36rem 0.6rem;
        font-size: 1.1rem;
        border: 0.5px solid
            var(--mdc-segmented-button-outline-color, rgba(255, 255, 255, 0.12));
        cursor: pointer;
    }

    input:focus {
        color: var(--clr-primary);
    }

    input::-webkit-calendar-picker-indicator {
        // Needed in order to hide the default icon
        color: transparent;
        background: none;
        z-index: 1;
        cursor: pointer;
    }

    input::before {
        // Shows the replacement icon, in this case the material icon for `event`
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
        transition: 0.1s;
    }

    #hint {
        font-size: 0.75rem;
        margin-left: 0.1rem;
        margin-top: 0.1rem;
        -webkit-font-smoothing: antialiased;
        display: block;
        opacity: 0;
        transition: opacity 150ms 0ms linear; // Animation properties ported from other mdc components
    }

    #hint.active {
        opacity: 1;
    }

    // Styles for an invalid input
    #hint.invalid {
        color: var(--clr-error);
        opacity: 1;
    }

    input.invalid {
        border-color: var(--clr-error);
        color: var(--clr-error);
    }
</style>
