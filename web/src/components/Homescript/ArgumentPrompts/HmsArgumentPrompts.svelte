<script lang="ts">
    import Dialog, { Content, Header, InitialFocus, Title } from '@smui/dialog'
    import { createEventDispatcher } from 'svelte'
    import Button, { Label } from '@smui/button'
    import Textfield from '@smui/textfield'
    import Switch from '@smui/switch'
    import type { homescriptArgData, homescriptArgSubmit } from '../../../homescript'
    import { createSnackbar } from '../../../global'
    import Progress from '../../Progress.svelte'
    import List, { Graphic, Item } from '@smui/list'
    import Radio from '@smui/radio'
    import FormField from '@smui/form-field'

    // Keeps track of whether the dialog should be open or not
    export let open = false

    // Event dispatcher
    const dispatch = createEventDispatcher()

    /*
        /// Important variables ////
        Are either bound to externally or frequently required internally
     */
    // Holds the argument list which is used to display the prompts
    // Is bound from other components to set up the prompts
    export let args: homescriptArgData[]

    // Saves the index of the argument which is currently shown as a prompt
    let currentArgumentIndex = 0

    // Represents the current argument at the `currentArgumentIndex` position in `args`
    let currentArg: homescriptArgData = {
        argKey: '',
        homescriptId: '',
        prompt: '',
        mdIcon: '',
        inputType: 'string',
        display: 'type_default',
    }
    // Update the `currentArg`
    $: if (currentArgumentIndex + 1 <= args.length) currentArg = args[currentArgumentIndex]

    /*
        //// Submit and next ////
        If the button is pressed the last time, the event dispatcher dispatches the 'submit' event.
        Then, the `argumentswithValues` is dispatched as the event detail
     */
    // Is produced when the final submit button is pressed
    // Is then submitted using the event dispatcher
    let argumentsWithValues: homescriptArgSubmit[] = []

    // Is called when the submit button is pressed
    function submit() {
        if (argumentsWithValues[currentArgumentIndex + 1] === undefined) {
            dispatch('submit', argumentsWithValues)
            currentArgumentIndex = 0
            open = false
            return
        }

        argumentsWithValues[currentArgumentIndex].key = currentArg.argKey

        // Reset all placeholders to their default value
        booleanPlaceholder = false
        numberPlaceholder = 0

        currentArgumentIndex++
    }

    /*
        //// Non-String binding and conversion ////
        Utility variables for non-string types with their conversion functions
        Will be converted to the `real` string representation using change listeners
    */
    // Placeholders for conversion
    let numberPlaceholder = 0
    let booleanPlaceholder = false

    // Conversion functions
    function updateFromNumber() {
        if (
            (currentArg.display === 'number_hour' || currentArg.display === 'number_minute') &&
            numberPlaceholder < 0
        )
            numberPlaceholder = 0
        if (currentArg.display === 'number_hour' && numberPlaceholder > 24) numberPlaceholder = 24
        if (currentArg.display === 'number_minute' && numberPlaceholder > 60) numberPlaceholder = 60

        argumentsWithValues[currentArgumentIndex].value = numberPlaceholder.toString()
    }

    function updateFromBoolean() {
        argumentsWithValues[currentArgumentIndex].value = booleanPlaceholder.toString()
    }

    // Change listeners to trigger conversion
    $: if (
        currentArg.inputType == 'number' &&
        argumentsWithValues.length > 0 &&
        numberPlaceholder !== undefined
    )
        updateFromNumber()

    $: if (
        currentArg.inputType === 'boolean' &&
        (currentArg.display === 'type_default' || currentArg.display === 'boolean_yes_no') &&
        argumentsWithValues.length > 0 &&
        booleanPlaceholder !== undefined
    )
        updateFromBoolean()

    $: if (
        currentArg.inputType === 'boolean' &&
        currentArg.display === 'boolean_on_off' &&
        argumentsWithValues.length > 0 &&
        booleanPlaceholder !== undefined
    )
        updateFromBoolean()

    /*
        //// Switches ////
        Used for when the `display` is set to `string_switches`
    */
    interface SwitchResponse {
        id: string
        name: string
        powerOn: boolean
        watts: number
    }

    // Switch variables
    let switchesLoaded = false
    let switches: SwitchResponse[] = []

    // Loads the user's personal switches
    async function loadSwitches() {
        try {
            const res = await (await fetch('/api/switch/list/personal')).json()
            if (res.success !== undefined && !res.success) throw Error(res.error)
            switches = res
            switchesLoaded = true
        } catch (err) {
            $createSnackbar(`Could not load switches: ${err}`)
        }
    }
    $: if (!switchesLoaded && currentArg.display === 'string_switches') loadSwitches()

    /*
        //// Initialization on dialog opening ////
       When the dialog is opened, create the `argumentsWithValues` list
    */
    $: if (open) createArgsWithValue()
    function createArgsWithValue() {
        for (let arg of args) argumentsWithValues.push({ key: arg.argKey, value: '' })
        if (args[0].inputType === 'boolean') {
            updateFromBoolean()
        }
    }
</script>

<Dialog
    bind:open
    aria-labelledby="title"
    aria-describedby="content"
    selection={(currentArg.display === 'string_switches' &&
        switchesLoaded &&
        switches.length > 0) ||
        (currentArg.inputType === 'boolean' &&
            (currentArg.display === 'type_default' || currentArg.display === 'boolean_yes_no'))}
>
    <Header>
        <Title id="title">{currentArg.prompt}</Title>
    </Header>
    <Content id="content">
        {#if argumentsWithValues.length > 0}
            <div
                class="inputs"
                class:centered={currentArg.display === 'number_hour' ||
                    currentArg.display === 'number_minute'}
            >
                {#if currentArg.inputType === 'string'}
                    {#if currentArg.display === 'type_default'}
                        <Textfield
                            style="width: 100%;"
                            bind:value={argumentsWithValues[currentArgumentIndex].value}
                            label={currentArg.argKey}
                        />
                    {:else if currentArg.display === 'string_switches'}
                        {#if switchesLoaded && switches.length === 0}
                            <span>No switches available.</span>
                            <br />
                            <span class="text-disabled">You can skip this prompt</span>
                        {:else if !switchesLoaded}
                            <Progress type="linear" loading={true} />
                        {:else}
                            <List radioList style="width: 100%;">
                                {#each switches as sw (sw.id)}
                                    <Item>
                                        <Graphic>
                                            <Radio
                                                bind:group={argumentsWithValues[
                                                    currentArgumentIndex
                                                ].value}
                                                value={sw.id}
                                            />
                                        </Graphic>
                                        <Label>
                                            {sw.name != '' ? sw.name : 'No Name'}
                                            <span class="text-disabled" style="font-size: .9rem;"
                                                >({sw.id})</span
                                            >
                                        </Label>
                                    </Item>
                                {/each}
                            </List>
                        {/if}
                    {/if}
                {:else if currentArg.inputType === 'number'}
                    {#if currentArg.display === 'type_default'}
                        <Textfield
                            style="width: 100%;"
                            bind:value={numberPlaceholder}
                            label={currentArg.argKey}
                            type="number"
                        />
                    {:else if currentArg.display === 'number_hour'}
                        <Textfield
                            style="width: 100%;"
                            bind:value={numberPlaceholder}
                            label={currentArg.argKey}
                            type="number"
                            min={0}
                            max={24}
                        />
                    {:else if currentArg.display === 'number_minute'}
                        <Textfield
                            style="width: 100%;"
                            bind:value={numberPlaceholder}
                            label={currentArg.argKey}
                            type="number"
                            min={0}
                            max={60}
                        />
                    {/if}
                {:else if currentArg.inputType === 'boolean'}
                    {#if currentArg.display === 'boolean_on_off'}
                        <br />
                        <FormField>
                            <Switch bind:checked={booleanPlaceholder} />
                            <span slot="label">{booleanPlaceholder ? 'On' : 'Off'}</span>
                        </FormField>
                    {:else}
                        <List radioList style="width: 100%;">
                            {#each [true, false] as opt (opt)}
                                <Item>
                                    <Graphic>
                                        <Radio bind:group={booleanPlaceholder} value={opt} />
                                    </Graphic>
                                    <Label>
                                        {#if currentArg.display === 'boolean_yes_no'}
                                            {opt ? 'Yes' : 'No'}
                                        {:else}
                                            {opt ? 'True' : 'False'}
                                        {/if}
                                    </Label>
                                </Item>
                            {/each}
                        </List>
                    {/if}
                {/if}
            </div>
        {/if}
        <div
            class="actions"
            class:selection={(currentArg.display === 'string_switches' &&
                switchesLoaded &&
                switches.length > 0) ||
                (currentArg.inputType === 'boolean' &&
                    (currentArg.display === 'type_default' ||
                        currentArg.display === 'boolean_yes_no'))}
        >
            <Button
                on:click={() => {
                    argumentsWithValues = []
                    currentArgumentIndex = 0
                    open = false
                }}
            >
                <Label>Cancel</Label>
            </Button>
            <Button use={[InitialFocus]} on:click={submit}>
                <Label>Submit</Label>
            </Button>
        </div>
    </Content>
</Dialog>

<style lang="scss">
    .inputs {
        height: 20rem;
        overflow: auto;

        &.centered {
            display: flex;
            justify-content: center;
        }
    }
    .actions {
        margin-top: 1rem;
        display: flex;
        justify-content: flex-end;

        &.selection {
            padding-right: 24px;
            padding-bottom: 20px;
        }
    }
</style>
