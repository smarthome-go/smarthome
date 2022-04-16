<script lang="ts">
    import Button,{ Label } from '@smui/button'
    import Dialog,{ Actions,Content,Header,Title } from '@smui/dialog'
    import FormField from '@smui/form-field'
    import IconButton from '@smui/icon-button'
    import Paper,{ Subtitle } from '@smui/paper'
    import Switch from '@smui/switch'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'
    import { data } from '../../global'

    let open = false

    export let username = ''
    export let forename = ''
    export let surname = ''
    export let darkTheme: boolean

    $: {
        if (username == $data.userData.user.username)
          $data.userData.user.darkTheme = darkTheme
    }
</script>

<Dialog bind:open fullscreen aria-labelledby="title" aria-describedby="content">
    <Header>
        <Title id="title">Manage User</Title>
        <IconButton action="close" class="material-icons">close</IconButton>
    </Header>
    <Content id="content">
        <div id="profile">
            <img
                class="mdc-elevation--z3"
                src={`/api/user/avatar/user/${username}`}
                alt=""
            />
            <div>
                <h6>{forename} {surname}</h6>
                <span>{username}</span>
            </div>
        </div>
        <h6 id="edit">Edit</h6>
        <div id="names">
            <div>
                <!-- Forename -->
                <Textfield
                    helperLine$style="width: 100%;"
                    label="Forename"
                    input$maxlength={30}
                    bind:value={forename}
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
        <div id="toggles" class="mdc-elevation--z1">
            <Paper variant="outlined">
                <Title>Toggles</Title>
                <Subtitle>Change theme and automation status</Subtitle>
                <div id="toggle-content">
                    <FormField>
                        <Switch bind:checked={darkTheme} />
                        <span slot="label">Dark Theme</span>
                      </FormField>
                </div>
            </Paper>
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
    #profile {
        display: flex;
        align-items: center;
        gap: 1rem;

        img {
            width: 5rem;
            height: 5rem;
            border-radius: 50%;
        }
    }
    h6 {
        margin: 0.5rem 0;
    }
    #edit {
        margin-top: 1rem;
    }

    #toggles {
        margin-top: 2rem;
        background-color: var(--clr-height-0-1);
    }
</style>
