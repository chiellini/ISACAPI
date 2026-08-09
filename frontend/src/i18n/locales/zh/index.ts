import landing from './landing'
import common from './common'
import dashboard from './dashboard'
import affiliate from './affiliate'
import channelMonitorV2 from './channelMonitorV2'
import batchImage from './batchImage'
import admin from './admin'
import misc from './misc'
import researchGroup from './researchGroup'
import provider from './provider'
import { zhCodexGuide } from '../codexGuide'

export default {
  ...landing,
  ...common,
  ...dashboard,
  ...affiliate,
  ...channelMonitorV2,
  ...batchImage,
  admin,
  ...misc,
  ...researchGroup,
  ...provider,
  codexGuide: zhCodexGuide,
}
