import landing from './landing'
import common from './common'
import dashboard from './dashboard'
import affiliate from './affiliate'
import batchImage from './batchImage'
import admin from './admin'
import misc from './misc'
import researchGroup from './researchGroup'
import provider from './provider'

export default {
  ...landing,
  ...common,
  ...dashboard,
  ...affiliate,
  ...batchImage,
  admin,
  ...misc,
  ...researchGroup,
  ...provider,
}
