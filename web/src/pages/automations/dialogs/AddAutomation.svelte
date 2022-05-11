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
    
    let selectedDays: string[] = ['mo']

    // BROKEN HERE
    $: console.log(selectedDays)

    let selectedHour = 0
    let selectedMinute = 0

    let selectedHms = ''
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content" fullscreen>
    <Header>
        <Title id="title">Add Automation</Title>
        <IconButton action="close" class="material-icons">close</IconButton>
    </Header>
    <Content id="content">
        <div class="container">
            <!-- Left -->
            <div class="left">
                <!-- Names and Text -->
                <div class="text">
                    <span class="text-hint"
                    >Name and description of the automation</span
                >
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

                <!-- Days -->
                <div class="days">
                    <span class="text-hint"
                        >Days on which the automation should run</span
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

                <!-- Time -->
                <div class="time">
                    <span class="text-hint">Time when the automation runs</span>
                    <TimePicker
                        bind:hour={selectedHour}
                        bind:minute={selectedMinute}
                        helperText={'Time'}
                        invalidText={'error'}
                    />
                </div>
            </div>

            <!-- Right -->
            <div class="right">
                <div class="hms">
                    <span class="text-hint">The Homescript to be executed</span>
                    <HmsSelector bind:selection={selectedHms} />
                </div>
            </div>
        </div>
    </Content>
    <Actions>
        <Button>
            <Label>Cancel</Label>
        </Button>
        <Button
            disabled={false}
            use={[InitialFocus]}
            on:click={() => {
                // Transform the selected days into data that the server understands
                data.days = selectedDays.map(d => days.indexOf(d))
                dispatch('add', data)
                // Reset values here
            }}
        >
            <Label>Create</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
    @use '../../../mixins' as *;
    .container {
        display: flex;
        flex-wrap: wrap;

        @include not-widescreen {
            flex-direction: column;
        }
    }

    .left,
    .right {
        @include widescreen {
            width: 50%;
            box-sizing: border-box;
            padding: 0 1rem;
        }
    }

    .days,
    .time {
        margin-top: 2rem;
    }

    .hms,
    .time,
    .days {
        display: flex;
        flex-direction: column;
        gap: 0.3rem;
    }

    .text {
        width: 90%;
    }
</style>
