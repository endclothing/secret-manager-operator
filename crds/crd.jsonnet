{
  apiVersion: 'apiextensions.k8s.io/v1beta1',
  kind: 'CustomResourceDefinition',
  metadata: {
    name: 'managedsecrets.endclothing.com',
  },
  spec: {
    group: 'endclothing.com',
    names: {
      kind: 'ManagedSecret',
      plural: 'managedsecrets',
      shortNames: [
        'ms',
      ],
      singular: 'managedsecret',
    },
    scope: 'Namespaced',
    validation: {
      openAPIV3Schema: {
        properties: {
          spec: {
            properties: {
              project: {
                type: 'string',
              },
              secret: {
                type: 'string',
              },
              generation: {
                type: 'integer',
              },
            },
            type: 'object',
          },
          status: {
            properties: {
              bucket: {
                type: 'string',
              },
              object: {
                type: 'string',
              },
              generation: {
                type: 'integer',
              },
              'error': {
                type: 'string',
              },
            },
            type: 'object',
          },
        },
        type: 'object',
      },
    },
    versions: [
      {
        name: 'v1',
        served: true,
        storage: true,
      },
    ],
  },
  status: {
    acceptedNames: {
      kind: '',
      plural: '',
    },
    conditions: [],
    storedVersions: [],
  },
}
