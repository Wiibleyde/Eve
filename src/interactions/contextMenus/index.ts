import * as quote from './message/quote';

import * as profilePicture from './user/profilePicture';
import * as banner from './user/banner';

const contextMessageMenus = {
    'Créer un citation': quote,
};

const contextUserMenus = {
    'Récupèrer la photo de profil': profilePicture,
    'Récupèrer la bannière': banner,
};

export { contextMessageMenus, contextUserMenus };
