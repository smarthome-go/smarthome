<script lang="ts">
    import SegmentedButton, { Label, Segment } from '@smui/segmented-button'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'
    import HelperText from '@smui/textfield/helper-text'
    import Page from '../../Page.svelte'
    import Reminder from './Reminder.svelte'

    // Add new inputs
    let inputName = ''
    let inputDescription = ''
    let selectedPriority = 'Normal'
</script>

<Page>
    <div id="content">
        <div id="container" class="mdc-elevation--z1">
            <Reminder
                id={1}
                name={'test'}
                description={'nein'}
                createdDate={'yesterday'}
                dueDate={'nein'}
                priority={0}
                userWasNotified={false}
            />
        </div>
        <div id="create" class="mdc-elevation--z1">
            <div id="name">
                <Textfield
                    style="width: 100%;"
                    helperLine$style="width: 100%;"
                    bind:value={inputName}
                    label="Name"
                    input$maxlength={100}
                >
                    <CharacterCounter slot="helper">0 / 100</CharacterCounter>
                </Textfield>
            </div>
            <div id="description">
                <Textfield
                    style="width: 100%;"
                    helperLine$style="width: 100%;"
                    textarea
                    bind:value={inputDescription}
                    label="Description"
                    input$rows={5}
                >
                    <HelperText slot="helper"
                        >Describe which task you want to accomplish</HelperText
                    >
                </Textfield>
            </div>
            <SegmentedButton
                segments={['Low', 'Normal', 'Medium', 'High', 'Urgent']}
                let:segment
                singleSelect
                bind:selected={selectedPriority}
            >
                <Segment {segment}>
                    <Label>{segment}</Label>
                </Segment>
            </SegmentedButton>
        </div>
    </div>
</Page>

<style lang="scss">
    @use '../../mixins' as *;
    #content {
        display: flex;
        flex-direction: column;
        @include widescreen {
            flex-direction: row;
            gap: 2rem;
        }
        margin: 1rem 1.5rem;
        gap: 1rem;
    }
    #container {
        background-color: var(--clr-height-0-1);
        border-radius: 0.4rem;
        padding: 1.5rem;
        @include widescreen {
            width: 50%;
        }
    }
    #create {
        background-color: var(--clr-height-0-1);
        border-radius: 0.4rem;
        padding: 1.5rem;
        @include widescreen {
            width: 50%;
        }
    }
    #description {
        margin-top: 1rem;

        :global(.mdc-text-field__resizer) { resize: none; }
    }
</style>
