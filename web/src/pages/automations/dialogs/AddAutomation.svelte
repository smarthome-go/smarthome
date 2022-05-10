<script lang="ts">
    import Button,{ Label } from '@smui/button'
    import Dialog,{
    Actions,
    Content,
    Header,
    InitialFocus,
    Title
    } from '@smui/dialog'
    import IconButton from '@smui/icon-button'
    import SegmentedButton,{ Segment } from '@smui/segmented-button'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'
    import { createEventDispatcher } from 'svelte'
    import TimePicker from '../../../components/TimePicker.svelte'
    import HmsSelector from '../HmsSelector.svelte'
    import type { addAutomation } from '../main'

    // Event dispatcher
    const dispatch = createEventDispatcher()

    let data: addAutomation = {
        days: [],
        description: '',
        enabled: true,
        homescriptId: '',
        hour: 0,
        minute: 0,
        name: '',
        timingMode: 'normal',
    }

    export let open = false

    const days: string[] = ['su', 'mo', 'tu', 'we', 'th', 'fr', 'sa']
    let selectedDays: string[] = []

    let selectedHour = 0
    let selectedMinute = 0

    let selectedHms = 'Tom Hanks'
    $: console.log(selectedHms)
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content" fullscreen>
    <Header>
        <Title id="title">Add Automation</Title>
        <IconButton action="close" class="material-icons">close</IconButton>
    </Header>
    <Content id="content">
        <div class="container">
            <div class="left">
                <Textfield
                    bind:value={data.name}
                    input$maxlength={1}
                    label="Name"
                    required
                >
                    <svelte:fragment slot="helper">
                        <CharacterCounter>0 / 1</CharacterCounter>
                    </svelte:fragment>
                </Textfield>
                <Textfield bind:value={data.description} label="Description" />

                <div class="days">
                    <span class="text-hint"
                        >Specifies on which days of the week the automation will
                        run.</span
                    >
                    <SegmentedButton
                        segments={days}
                        let:segment
                        bind:selectedDays
                    >
                        <Segment {segment}>
                            <Label>{segment}</Label>
                        </Segment>
                    </SegmentedButton>
                </div>
                <div class="time">
                    <span class="text-hint"
                        >The time on which the automation will run</span
                    >
                    <TimePicker
                        bind:hour={selectedHour}
                        bind:minute={selectedMinute}
                        helperText={'Time'}
                        invalidText={'error'}
                    />
                </div>

                <!-- List
                <div class="list">
                    <Select bind:value={selectedHms} label="Select Menu">
                        {#each ['a', 'b'] as selectedHms}
                            <Option value={selectedHms}>{selectedHms}</Option>
                        {/each}
                    </Select>
                    <pre class="status">Selected: {selectedHms}</pre>
                </div> -->
            </div>
            <div class="right">
                <HmsSelector bind:selection={selectedHms} />
            </div>
        </div>
    </Content>
    <Actions>
        <Button>
            <Label>Cancel</Label>
        </Button>
        <Button
            disabled={true}
            use={[InitialFocus]}
            on:click={() => {
                dispatch('add', data)
                // Reset values here
            }}
        >
            <Label>Create</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
    .days,
    .time {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }
    .time {
        display: flex;
        margin-top: 1.5rem;
    }
    .container {
        display: flex;
        justify-content: space-between;
        flex-wrap: wrap;
    }
</style>
