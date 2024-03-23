/*
** ssl.c
** Simple String List
** array of characters or zero terminated strings
*/

#include <stdlib.h>
#include <string.h>

#include "ssl.h"


struct pcl {
  char *pch;
  unsigned int len;
};

struct _SSL
{
  int m;           // Taille de la liste
  int n;           // Nombre de chaînes référencées
  struct pcl *sl;  // Liste des pointeurs de chaînes
};

SSL SSL_new(unsigned int size)
{
  SSL ssl;

  if ( size == 0 ) size = 8;

  ssl = (SSL)malloc(sizeof(struct _SSL));

  if ( ssl != NULL )
  {
    ssl->m = size;
    ssl->n = 0;
    if ( ( ssl->sl=(struct pcl*)calloc(size, sizeof(struct pcl))) == NULL ) return SSL_del(ssl);
  }

  return ssl;
}

SSL SSL_del(SSL ssl)
{
  if ( ssl != NULL )
  {
    for ( int i = 0; i < ssl->n; i++ ) // Destruction de chaque chaîne
    {
      if ( ssl->sl[i].pch != NULL ) free(ssl->sl[i].pch);
    }
    if ( ssl->sl != NULL ) free(ssl->sl);
    free(ssl);
  }

  return NULL;
}

unsigned int SSL_count(SSL ssl)
{
  if ( ssl != NULL )
    return ssl->n;
  else
    return 0;
}

unsigned int SSL_add(SSL ssl, char *sz) // Zero terminated String (0t)
{
  return SSL_add2(ssl, sz, strlen(sz));
}

unsigned int SSL_add2(SSL ssl, char *str, unsigned int len) // Array of characters (or zero terminated string)
{
  if ( ssl != NULL )
  {
    // Adapter la liste - pas assez de place
    if ( ssl->n >= ssl->m )
    {
      struct pcl *oldsl = ssl->sl; // Sauver l'emplacement de l'ancienne liste
      unsigned int size = 2 * ssl->m; // Allouer 2 fois plus pour rester efficace dans des appels répétés
      struct pcl *newsl = (struct pcl*)calloc(size, sizeof(struct pcl)); // Allouer une nouvelle liste
      if ( newsl == NULL ) return 0; // Préserver l'état actuel en sortant sans modification
      ssl->m = size; // Sinon actualiser l'objet en fonction de ces nouvelles valeurs
      ssl->sl = newsl;
      memcpy(ssl->sl, oldsl, ssl->n * sizeof(struct pcl)); // Y recopier l'ancienne liste
      free(oldsl); // Désallouer l'ancienne liste
    }
    // Allouer et recopier la substring ss
    ssl->sl[ssl->n].pch = (char*)calloc(len+1, sizeof(char)); // Y compris un 0t
    if ( ssl->sl[ssl->n].pch == NULL ) return 0;
    memcpy(ssl->sl[ssl->n].pch, str, len);
    ssl->sl[ssl->n].len = len;
    return ++(ssl->n); // Tout est OK
  }
  else
    return 0;
}

char *SSL_get(SSL ssl, int index)
{
  static char nul[] = "nul";
  
  if ( ssl != NULL && index >= 0 && index < ssl->n ) // index in [0, n-1]
    return ssl->sl[index].pch;
  else
    return nul;
}

unsigned int SSL_len(SSL ssl, int index)
{
  if ( ssl != NULL && index >= 0 && index < ssl->n ) // index in [0, n-1]
    return ssl->sl[index].len;
  else
    return 0;
}

/* eof */
