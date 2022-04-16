<script lang="ts">
    import Button,{ Label } from '@smui/button'
    import Dialog,{ Actions,Content,Header,Title } from '@smui/dialog'
    import IconButton from '@smui/icon-button'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'

    let open = false

    export let username = ''
    export let forename = ''
    export let surname = ''

    export let title = 'Manage User'
</script>

<Dialog bind:open fullscreen aria-labelledby="title" aria-describedby="content">
    <Header>
        <Title id="title">{title}</Title>
        <IconButton action="close" class="material-icons">close</IconButton>
    </Header>
    <Content id="content">
        <div id="profile">
            <img src={`/api/user/avatar/user/${username}`} alt="">
            <div>
                <h6>{forename} {surname}</h6>
            </div>
        </div>
        <div id="names">
            <div>
                <!-- Forename -->
                <Textfield
                    helperLine$style="width: 100%;"
                    label="Forename"
                    input$maxlength={30}
                    bind:value={surname}
                >
                    <CharacterCounter slot="helper">0 / 30</CharacterCounter>
                </Textfield>
            </div>
            <div>
                <!-- Surname -->
                <Textfield
                    helperLine$style="width: 100%;"
                    label="Surname"
                    input$maxlength={30}
                    bind:value={surname}
                >
                    <CharacterCounter slot="helper">0 / 30</CharacterCounter>
                </Textfield>
            </div>
        </div>
    </Content>
    <Actions>
        <Button defaultAction>
            <Label>Save</Label>
        </Button>
        <Button>
            <Label>Cancel</Label>
        </Button>
    </Actions>
</Dialog>

<IconButton
    class="material-icons"
    on:click={async () => {
        open = true
    }}
    title="Manage">edit</IconButton
>

<style lang="scss">
    #names {
        display: flex;
        gap: 2rem;
    }
</style>
