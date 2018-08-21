// @flow
import * as _Avatar from '../common-adapters/avatar'
import * as _Usernames from '../common-adapters/usernames'
import * as _WaitingButton from '../common-adapters/waiting-button'
import * as _TeamDropdownMenu from '../chat/conversation/info-panel/menu/container'
import * as _CopyText from '../common-adapters/copy-text'
import {action} from '@storybook/addon-actions'
import {createPropProvider} from './storybook.shared'

/*
 * Some common prop factory creators.
 *
 *  Params: specific pieces of data (not necessarily store data)
 *          that are needed to derive view props
 *  Output: a map of DisplayName: Function(...) that returns the
 *          view props the connected component is concerned with
 */

const defaultYou = 'ayoubd'
const defaultFollowing = ['max', 'cnojima', 'cdixon']
const defaultFollowers = ['max', 'akalin']

export const Usernames = (following: string[] = defaultFollowing, you: string = defaultYou) => ({
  Usernames: (ownProps: _Usernames.ConnectedProps): _Usernames.Props => {
    const {usernames} = ownProps
    const users = (usernames || [])
      .map(username => ({username, following: following.includes(username), you: username === you}))
      .filter(u => !ownProps.skipSelf || !u.you)

    return {
      ...ownProps,
      users,
      onUsernameClicked: action('onUsernameClicked'),
    }
  },
})

export const WaitingButton = () => ({
  WaitingButton: (ownProps: _WaitingButton.OwnProps): _WaitingButton.Props => ({
    ...ownProps,
    storeWaiting: false,
  }),
})

export const Avatar = (following: string[] = defaultFollowing, followers: string[] = defaultFollowers) => ({
  Avatar: (ownProps: _Avatar.Props) => _Avatar.mockOwnToViewProps(ownProps, following, followers, action),
})

export const TeamDropdownMenu = (adminTeams?: string[], teamMemberCounts?: {[key: string]: number}) => ({
  TeamDropdownMenu: (ownProps: _TeamDropdownMenu.OwnProps): _TeamDropdownMenu.Props => ({
    _loadOperations: action('_loadOperations'),
    attachTo: ownProps.attachTo,
    badgeSubscribe: false,
    canAddPeople: (adminTeams && adminTeams.includes(ownProps.teamname)) || true,
    isSmallTeam: ownProps.isSmallTeam,
    manageChannelsSubtitle: ownProps.isSmallTeam ? 'Turns this into a big team' : '',
    manageChannelsTitle: ownProps.isSmallTeam ? 'Create chat channels...' : 'Manage chat channels',
    memberCount: (teamMemberCounts && teamMemberCounts[ownProps.teamname]) || 100,
    teamname: ownProps.teamname,
    visible: ownProps.visible,
    onAddPeople: action('onAddPeople'),
    onHidden: ownProps.onHidden,
    onInvite: action('onInvite'),
    onLeaveTeam: action('onLeaveTeam'),
    onManageChannels: action('onManageChannels'),
    onViewTeam: action('onViewTeam'),
  }),
})

const CopyText = () => ({
  CopyText: (p: _CopyText.Props) => ({...p, copyToClipboard: action('copyToClipboard')}),
})

export const Common = () => ({
  ...Usernames(),
  ...Avatar(),
  ...WaitingButton(),
  ...CopyText(),
})

export const createPropProviderWithCommon = (custom: ?Object) =>
  createPropProvider({
    ...Common(),
    ...(custom || {}),
  })
